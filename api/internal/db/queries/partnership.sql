-- name: CreatePartnership :exec
INSERT INTO partnerships (id, consignor_org_id, store_org_id, status, started_at, created_at, updated_at)
VALUES (
    @id::uuid, @consignor_org_id::uuid, @store_org_id::uuid, 'ACTIVE', @started_at::timestamptz,
    @started_at::timestamptz, @started_at::timestamptz
);

-- name: ListPartnerships :many
SELECT
    p.id, p.status, p.started_at, COALESCE(p.ended_at, p.started_at)::timestamptz AS ended_at,
    (p.ended_at IS NOT NULL)::boolean AS has_ended,
    o.id AS counterpart_id, o.name AS counterpart_name, o.type AS counterpart_type,
    COALESCE((SELECT SUM(b.balance) FROM batch_balances AS b WHERE b.partnership_id = p.id), 0)::bigint AS stock_qty,
    COALESCE((
        SELECT SUM(st.consignor_share_amount) FROM settlements AS st
        WHERE st.partnership_id = p.id AND st.status = 'ISSUED'
    ), 0)::bigint AS unpaid_amount
FROM partnerships AS p
JOIN organizations AS o
    ON o.id = CASE WHEN p.consignor_org_id = @organization_id::uuid THEN p.store_org_id ELSE p.consignor_org_id END
WHERE (p.consignor_org_id = @organization_id::uuid OR p.store_org_id = @organization_id::uuid)
    AND (@status::text = '' OR p.status = @status::text)
    AND (@from_start::boolean OR (p.started_at, p.id) < (@after_at::timestamptz, @after_id::uuid))
ORDER BY p.started_at DESC, p.id DESC
LIMIT @row_limit::bigint;

-- name: GetPartnership :one
SELECT
    p.id, p.status, p.started_at, COALESCE(p.ended_at, p.started_at)::timestamptz AS ended_at,
    (p.ended_at IS NOT NULL)::boolean AS has_ended, p.show_on_public, p.end_reason,
    o.id AS counterpart_id, o.type AS counterpart_type, o.name AS counterpart_name, o.slug AS counterpart_slug,
    o.description, o.whatsapp, o.city_name, o.district, o.address,
    (o.location IS NOT NULL)::boolean AS has_location,
    COALESCE(ST_Y(o.location::geometry), 0)::float8 AS latitude,
    COALESCE(ST_X(o.location::geometry), 0)::float8 AS longitude,
    COALESCE(ST_Y(o.location_public::geometry), 0)::float8 AS public_latitude,
    COALESCE(ST_X(o.location_public::geometry), 0)::float8 AS public_longitude,
    COALESCE(f.storage_key, '')::text AS logo_storage_key
FROM partnerships AS p
JOIN organizations AS o
    ON o.id = CASE WHEN p.consignor_org_id = @organization_id::uuid THEN p.store_org_id ELSE p.consignor_org_id END
LEFT JOIN files AS f ON f.id = o.logo_file_id
WHERE p.id = @id::uuid AND (p.consignor_org_id = @organization_id::uuid OR p.store_org_id = @organization_id::uuid);

-- name: ListPartnerCatalog :many
SELECT
    pr.id, pr.name, pr.sku, pr.unit, pr.description, pr.net_content, c.name AS category_name,
    COALESCE(f.storage_key, '')::text AS photo_storage_key, COALESCE(pr.shelf_life_days, 0)::bigint AS shelf_life_days,
    (ai.product_id IS NOT NULL)::boolean AS is_agreed, COALESCE(ai.is_active, false)::boolean AS is_active,
    COALESCE(ai.pricing_model, '')::text AS pricing_model, COALESCE(ai.retail_price, 0)::bigint AS retail_price,
    COALESCE(ai.commission_bps, 0)::bigint AS commission_bps, COALESCE(ai.deposit_price, 0)::bigint AS deposit_price
FROM partnerships AS p
JOIN products AS pr ON pr.consignor_org_id = p.consignor_org_id
JOIN categories AS c ON c.id = pr.category_id
LEFT JOIN files AS f ON f.id = pr.photo_file_id
LEFT JOIN agreements AS a ON a.partnership_id = p.id AND a.status = 'ACTIVE'
LEFT JOIN agreement_items AS ai ON ai.agreement_id = a.id AND ai.product_id = pr.id
WHERE p.id = @partnership_id::uuid
    AND (p.consignor_org_id = @organization_id::uuid OR p.store_org_id = @organization_id::uuid)
    AND (NOT pr.is_archived OR ai.product_id IS NOT NULL)
ORDER BY (ai.product_id IS NULL), pr.name, pr.id;

-- name: LockPartnershipForEnding :one
SELECT p.id, p.status, p.consignor_org_id, c.name AS consignor_name, p.store_org_id, s.name AS store_name
FROM partnerships AS p
JOIN organizations AS c ON c.id = p.consignor_org_id
JOIN organizations AS s ON s.id = p.store_org_id
WHERE p.id = @id::uuid AND (p.consignor_org_id = @organization_id::uuid OR p.store_org_id = @organization_id::uuid)
FOR UPDATE OF p;

-- name: ListPartnershipBlockers :many
SELECT blockers.kind, blockers.document_id, blockers.document_number, blockers.qty, blockers.amount
FROM (
    SELECT
        'STOCK_AT_STORE'::text AS kind, ''::text AS document_id, ''::text AS document_number,
        COALESCE(SUM(b.balance), 0)::bigint AS qty, 0::bigint AS amount
    FROM batch_balances AS b
    WHERE b.partnership_id = @partnership_id::uuid
    HAVING COALESCE(SUM(b.balance), 0) > 0
    UNION ALL
    SELECT 'PENDING_DROP'::text, d.id::text, d.number, 0::bigint, 0::bigint
    FROM drops AS d
    WHERE d.partnership_id = @partnership_id::uuid AND d.status = 'RECORDED'
        AND d.received_at > @confirm_cutoff::timestamptz
    UNION ALL
    SELECT 'PENDING_COUNT'::text, c.id::text, c.number, 0::bigint, 0::bigint
    FROM counts AS c
    WHERE c.partnership_id = @partnership_id::uuid AND c.status = 'RECORDED'
        AND c.received_at > @confirm_cutoff::timestamptz
    UNION ALL
    SELECT 'PENDING_PAYMENT'::text, st.id::text, st.number, 0::bigint, p.amount
    FROM payments AS p
    JOIN settlements AS st ON st.id = p.settlement_id
    WHERE p.partnership_id = @partnership_id::uuid AND p.status = 'PENDING_CONFIRMATION'
    UNION ALL
    SELECT 'UNPAID_SETTLEMENT'::text, st.id::text, st.number, 0::bigint, st.consignor_share_amount
    FROM settlements AS st
    WHERE st.partnership_id = @partnership_id::uuid AND st.status = 'ISSUED'
        AND NOT EXISTS (
            SELECT 1 FROM payments AS p WHERE p.settlement_id = st.id AND p.status = 'PENDING_CONFIRMATION'
        )
) AS blockers
ORDER BY blockers.kind, blockers.document_number;

-- name: EndPartnership :execrows
UPDATE partnerships
SET status = 'ENDED', ended_at = @ended_at::timestamptz, ended_by_org_id = @ended_by_org_id::uuid,
    end_reason = @end_reason::text, updated_at = @ended_at::timestamptz
WHERE id = @id::uuid AND status = 'ACTIVE';

-- name: SetPartnershipShownOnPublic :execrows
UPDATE partnerships
SET show_on_public = @show_on_public::boolean, updated_at = @updated_at::timestamptz
WHERE id = @id::uuid AND store_org_id = @store_org_id::uuid;
