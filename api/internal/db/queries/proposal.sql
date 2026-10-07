-- name: NextDocumentNumber :one
INSERT INTO document_sequences (organization_id, prefix, period, last_value)
VALUES (@organization_id::uuid, @prefix::text, @period::text, 1)
ON CONFLICT (organization_id, prefix, period) DO UPDATE
SET last_value = document_sequences.last_value + 1
RETURNING last_value;

-- name: CreateProposal :exec
INSERT INTO proposals (
    id, number, kind, consignor_org_id, store_org_id, from_org_id, awaiting_org_id, partnership_id, invite_code,
    status, current_revision_no, expires_at, created_at, updated_at
) VALUES (
    @id::uuid, @number::text, @kind::text, NULLIF(@consignor_org_id::text, '')::uuid,
    NULLIF(@store_org_id::text, '')::uuid, @from_org_id::uuid, NULLIF(@awaiting_org_id::text, '')::uuid,
    NULLIF(@partnership_id::text, '')::uuid, NULLIF(@invite_code::text, ''), 'OPEN', 1,
    @expires_at::timestamptz, @created_at::timestamptz, @created_at::timestamptz
);

-- name: CreateProposalRevision :exec
INSERT INTO proposal_revisions (
    id, proposal_id, revision_no, author_org_id, author_user_id, settlement_schedule, payment_terms_days,
    count_rhythm, notes, message, created_at
) VALUES (
    @id::uuid, @proposal_id::uuid, @revision_no::bigint, @author_org_id::uuid, @author_user_id::uuid,
    @settlement_schedule::text, @payment_terms_days::bigint, @count_rhythm::text, @notes::text,
    @message::text, @created_at::timestamptz
);

-- name: AddProposalItems :execrows
INSERT INTO proposal_items (
    revision_id, product_id, action, pricing_model, retail_price, commission_bps, deposit_price
)
SELECT @revision_id::uuid, item.product_id, item.action, item.pricing_model, item.retail_price,
    item.commission_bps, item.deposit_price
FROM (
    SELECT
        unnest(@product_ids::uuid[]) AS product_id, unnest(@actions::text[]) AS action,
        unnest(@pricing_models::text[]) AS pricing_model, unnest(@retail_prices::bigint[]) AS retail_price,
        unnest(@commission_bps::bigint[]) AS commission_bps, unnest(@deposit_prices::bigint[]) AS deposit_price
) AS item
JOIN products AS p ON p.id = item.product_id
WHERE p.consignor_org_id = @consignor_org_id::uuid AND (NOT p.is_archived OR item.action <> 'ADD');

-- name: GetProposal :one
SELECT
    p.id, p.number, p.kind, p.status, p.from_org_id, COALESCE(p.awaiting_org_id::text, '')::text AS awaiting_org_id,
    COALESCE(p.invite_code, '')::text AS invite_code, p.current_revision_no, p.expires_at, p.created_at,
    COALESCE(p.consignor_org_id::text, '')::text AS consignor_org_id, COALESCE(c.name, '')::text AS consignor_name,
    COALESCE(p.store_org_id::text, '')::text AS store_org_id, COALESCE(s.name, '')::text AS store_name,
    r.id AS revision_id, r.author_org_id, r.author_user_id, r.settlement_schedule, r.payment_terms_days,
    r.count_rhythm, r.notes, r.message, r.created_at AS revised_at, p.decision_reason,
    COALESCE(p.partnership_id::text, '')::text AS partnership_id
FROM proposals AS p
JOIN proposal_revisions AS r ON r.proposal_id = p.id AND r.revision_no = p.current_revision_no
LEFT JOIN organizations AS c ON c.id = p.consignor_org_id
LEFT JOIN organizations AS s ON s.id = p.store_org_id
WHERE p.id = @id::uuid AND (p.consignor_org_id = @organization_id::uuid OR p.store_org_id = @organization_id::uuid);

-- name: LockProposal :one
SELECT
    p.id, p.number, p.kind, p.status, p.from_org_id, COALESCE(p.awaiting_org_id::text, '')::text AS awaiting_org_id,
    COALESCE(p.invite_code, '')::text AS invite_code, p.current_revision_no, p.expires_at, p.created_at,
    COALESCE(p.consignor_org_id::text, '')::text AS consignor_org_id, COALESCE(c.name, '')::text AS consignor_name,
    COALESCE(p.store_org_id::text, '')::text AS store_org_id, COALESCE(s.name, '')::text AS store_name,
    r.id AS revision_id, r.author_org_id, r.author_user_id, r.settlement_schedule, r.payment_terms_days,
    r.count_rhythm, r.notes, r.message, r.created_at AS revised_at, p.decision_reason,
    COALESCE(p.partnership_id::text, '')::text AS partnership_id
FROM proposals AS p
JOIN proposal_revisions AS r ON r.proposal_id = p.id AND r.revision_no = p.current_revision_no
LEFT JOIN organizations AS c ON c.id = p.consignor_org_id
LEFT JOIN organizations AS s ON s.id = p.store_org_id
WHERE p.id = @id::uuid AND (p.consignor_org_id = @organization_id::uuid OR p.store_org_id = @organization_id::uuid)
FOR UPDATE OF p;

-- name: ListProposalItems :many
SELECT
    i.product_id, pr.name AS product_name, pr.sku, pr.unit, i.action, i.pricing_model, i.retail_price,
    i.commission_bps, i.deposit_price
FROM proposal_items AS i
JOIN products AS pr ON pr.id = i.product_id
WHERE i.revision_id = @revision_id::uuid
ORDER BY pr.name, pr.id;

-- name: GetInvitation :one
SELECT
    p.id, p.status, p.expires_at, p.from_org_id, o.name AS sender_name, o.type AS sender_type,
    COALESCE(p.consignor_org_id::text, '')::text AS consignor_org_id,
    COALESCE(p.store_org_id::text, '')::text AS store_org_id,
    (SELECT count(*) FROM proposal_items AS i JOIN proposal_revisions AS r ON r.id = i.revision_id
        WHERE r.proposal_id = p.id AND r.revision_no = p.current_revision_no)::bigint AS item_count
FROM proposals AS p
JOIN organizations AS o ON o.id = p.from_org_id
WHERE p.invite_code = @invite_code::text;

-- name: ClaimInvitation :execrows
UPDATE proposals
SET consignor_org_id = COALESCE(consignor_org_id, NULLIF(@consignor_org_id::text, '')::uuid),
    store_org_id = COALESCE(store_org_id, NULLIF(@store_org_id::text, '')::uuid),
    awaiting_org_id = @claimer_org_id::uuid,
    updated_at = @claimed_at::timestamptz
WHERE id = @id::uuid AND status = 'OPEN' AND awaiting_org_id IS NULL;

-- name: HasActivePartnership :one
SELECT EXISTS (
    SELECT 1 FROM partnerships
    WHERE consignor_org_id = @consignor_org_id::uuid AND store_org_id = @store_org_id::uuid AND status = 'ACTIVE'
)::boolean;

-- name: DecideProposal :execrows
UPDATE proposals
SET status = @status::text,
    decided_by = @decided_by::uuid,
    decided_at = @decided_at::timestamptz,
    decision_reason = @decision_reason::text,
    partnership_id = COALESCE(NULLIF(@partnership_id::text, '')::uuid, partnership_id),
    updated_at = @decided_at::timestamptz
WHERE id = @id::uuid AND status = 'OPEN' AND current_revision_no = @revision_no::bigint;

-- name: ListProposals :many
SELECT
    p.id, p.number, p.kind, p.status, p.from_org_id, COALESCE(p.awaiting_org_id::text, '')::text AS awaiting_org_id,
    COALESCE(CASE WHEN p.consignor_org_id = @organization_id::uuid THEN s.name ELSE c.name END, '')::text
        AS counterpart_name,
    p.expires_at, p.created_at
FROM proposals AS p
LEFT JOIN organizations AS c ON c.id = p.consignor_org_id
LEFT JOIN organizations AS s ON s.id = p.store_org_id
WHERE (p.consignor_org_id = @organization_id::uuid OR p.store_org_id = @organization_id::uuid)
    AND ((p.from_org_id = @organization_id::uuid) = @outgoing::boolean)
    AND (@from_start::boolean OR (p.created_at, p.id) < (@after_at::timestamptz, @after_id::uuid))
ORDER BY p.created_at DESC, p.id DESC
LIMIT @row_limit::bigint;
