-- name: CreateAgreement :exec
INSERT INTO agreements (
    id, partnership_id, version, status, settlement_schedule, payment_terms_days, count_rhythm, notes,
    source_proposal_id, proposed_by, accepted_by, accepted_at, effective_from, created_at
) VALUES (
    @id::uuid, @partnership_id::uuid, @version::bigint, 'ACTIVE', @settlement_schedule::text,
    @payment_terms_days::bigint, @count_rhythm::text, @notes::text, @source_proposal_id::uuid, @proposed_by::uuid,
    @accepted_by::uuid, @accepted_at::timestamptz, @accepted_at::timestamptz, @accepted_at::timestamptz
);

-- name: AddAgreementItems :exec
INSERT INTO agreement_items (
    agreement_id, product_id, pricing_model, retail_price, commission_bps, deposit_price, is_active
)
SELECT @agreement_id::uuid, item.product_id, item.pricing_model, item.retail_price, item.commission_bps,
    item.deposit_price, item.is_active
FROM (
    SELECT
        unnest(@product_ids::uuid[]) AS product_id, unnest(@pricing_models::text[]) AS pricing_model,
        unnest(@retail_prices::bigint[]) AS retail_price, unnest(@commission_bps::bigint[]) AS commission_bps,
        unnest(@deposit_prices::bigint[]) AS deposit_price, unnest(@active_flags::boolean[]) AS is_active
) AS item;

-- name: GetAgreementVersion :one
SELECT
    p.status AS partnership_status,
    a.id, a.partnership_id, a.version, a.status, a.settlement_schedule, a.payment_terms_days, a.count_rhythm,
    a.notes, a.source_proposal_id, pr.number AS proposal_number, a.proposed_by, a.accepted_by, a.accepted_at,
    pu.full_name AS proposer_name, COALESCE(po.name, '')::text AS proposer_org_name,
    au.full_name AS acceptor_name, COALESCE(ao.name, '')::text AS acceptor_org_name,
    p.consignor_org_id, c.name AS consignor_name, p.store_org_id, s.name AS store_name
FROM agreements AS a
JOIN partnerships AS p ON p.id = a.partnership_id
JOIN proposals AS pr ON pr.id = a.source_proposal_id
JOIN organizations AS c ON c.id = p.consignor_org_id
JOIN organizations AS s ON s.id = p.store_org_id
JOIN users AS pu ON pu.id = a.proposed_by
JOIN users AS au ON au.id = a.accepted_by
LEFT JOIN memberships AS pm ON pm.user_id = a.proposed_by
LEFT JOIN organizations AS po ON po.id = pm.organization_id
LEFT JOIN memberships AS am ON am.user_id = a.accepted_by
LEFT JOIN organizations AS ao ON ao.id = am.organization_id
WHERE a.partnership_id = @partnership_id::uuid
    AND ((@version::bigint = 0 AND a.status = 'ACTIVE') OR a.version = @version::bigint)
    AND (p.consignor_org_id = @organization_id::uuid OR p.store_org_id = @organization_id::uuid);

-- name: ListAgreementItems :many
SELECT
    i.product_id, pr.name AS product_name, pr.sku, pr.unit, i.pricing_model, i.retail_price, i.commission_bps,
    i.deposit_price, i.is_active
FROM agreement_items AS i
JOIN products AS pr ON pr.id = i.product_id
WHERE i.agreement_id = @agreement_id::uuid
ORDER BY pr.name, pr.id;

-- name: LockActiveAgreement :one
SELECT
    a.id, a.partnership_id, a.version, a.settlement_schedule, a.payment_terms_days, a.count_rhythm, a.notes,
    a.source_proposal_id, pr.number AS proposal_number, a.proposed_by, a.accepted_by, a.accepted_at
FROM agreements AS a
JOIN proposals AS pr ON pr.id = a.source_proposal_id
WHERE a.partnership_id = @partnership_id::uuid AND a.status = 'ACTIVE'
FOR UPDATE OF a;

-- name: SupersedeAgreement :execrows
UPDATE agreements SET status = 'SUPERSEDED' WHERE id = @id::uuid AND status = 'ACTIVE';

-- name: ListAgreementVersions :many
SELECT
    a.id, a.version, a.status, a.settlement_schedule, a.payment_terms_days, a.count_rhythm, a.notes,
    pr.number AS proposal_number, a.accepted_at
FROM agreements AS a
JOIN partnerships AS p ON p.id = a.partnership_id
JOIN proposals AS pr ON pr.id = a.source_proposal_id
WHERE a.partnership_id = @partnership_id::uuid
    AND (p.consignor_org_id = @organization_id::uuid OR p.store_org_id = @organization_id::uuid)
ORDER BY a.version;

-- name: ListPartnershipAgreementItems :many
SELECT
    i.agreement_id, i.product_id, pr.name AS product_name, pr.sku, pr.unit, i.pricing_model, i.retail_price,
    i.commission_bps, i.deposit_price, i.is_active
FROM agreement_items AS i
JOIN agreements AS a ON a.id = i.agreement_id
JOIN products AS pr ON pr.id = i.product_id
WHERE a.partnership_id = @partnership_id::uuid
ORDER BY a.version, pr.name, pr.id;
