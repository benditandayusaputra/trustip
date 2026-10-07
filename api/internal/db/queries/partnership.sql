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
    o.id AS counterpart_id, o.name AS counterpart_name, o.type AS counterpart_type
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
    (p.ended_at IS NOT NULL)::boolean AS has_ended,
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
    COALESCE(f.storage_key, '')::text AS photo_storage_key,
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
