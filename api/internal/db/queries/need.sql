-- name: CreateNeed :execrows
INSERT INTO needs (
    id, store_org_id, title, category_id, description, est_qty_per_week, max_retail_price, preferred_model,
    offered_commission_bps, status, expires_at, created_by, created_at, updated_at
)
SELECT
    @id::uuid, @store_org_id::uuid, @title::text, c.id, @description::text, sqlc.narg(est_qty_per_week)::bigint,
    sqlc.narg(max_retail_price)::bigint, NULLIF(@preferred_model::text, ''),
    sqlc.narg(offered_commission_bps)::bigint, 'OPEN', @expires_at::timestamptz, @created_by::uuid,
    @created_at::timestamptz, @created_at::timestamptz
FROM categories AS c
WHERE c.id = @category_id::uuid;

-- name: GetNeed :one
SELECT
    n.id, n.store_org_id, o.name AS store_name, n.title, n.category_id, c.name AS category_name, n.description,
    n.est_qty_per_week, n.max_retail_price, COALESCE(n.preferred_model, '')::text AS preferred_model,
    n.offered_commission_bps, n.status, n.expires_at, n.created_at,
    (SELECT count(*) FROM proposals AS p WHERE p.need_id = n.id)::bigint AS response_count
FROM needs AS n
JOIN organizations AS o ON o.id = n.store_org_id
JOIN categories AS c ON c.id = n.category_id
WHERE n.id = @id::uuid AND n.store_org_id = @store_org_id::uuid;

-- name: LockNeed :one
SELECT
    n.id, n.store_org_id, o.name AS store_name, n.title, n.category_id, c.name AS category_name, n.description,
    n.est_qty_per_week, n.max_retail_price, COALESCE(n.preferred_model, '')::text AS preferred_model,
    n.offered_commission_bps, n.status, n.expires_at, n.created_at,
    (SELECT count(*) FROM proposals AS p WHERE p.need_id = n.id)::bigint AS response_count
FROM needs AS n
JOIN organizations AS o ON o.id = n.store_org_id
JOIN categories AS c ON c.id = n.category_id
WHERE n.id = @id::uuid AND n.store_org_id = @store_org_id::uuid
FOR UPDATE OF n;

-- name: ListStoreNeeds :many
SELECT
    n.id, n.store_org_id, o.name AS store_name, n.title, n.category_id, c.name AS category_name, n.description,
    n.est_qty_per_week, n.max_retail_price, COALESCE(n.preferred_model, '')::text AS preferred_model,
    n.offered_commission_bps, n.status, n.expires_at, n.created_at,
    (SELECT count(*) FROM proposals AS p WHERE p.need_id = n.id)::bigint AS response_count
FROM needs AS n
JOIN organizations AS o ON o.id = n.store_org_id
JOIN categories AS c ON c.id = n.category_id
WHERE n.store_org_id = @store_org_id::uuid
    AND (@status::text = '' OR n.status = @status::text)
    AND (@from_start::boolean OR (n.created_at, n.id) < (@after_at::timestamptz, @after_id::uuid))
ORDER BY n.created_at DESC, n.id DESC
LIMIT @row_limit::bigint;

-- name: SaveNeedChange :execrows
UPDATE needs
SET status = @status::text,
    reminded_at = CASE WHEN @expires_at::timestamptz <> expires_at THEN NULL ELSE reminded_at END,
    expires_at = @expires_at::timestamptz,
    updated_at = @changed_at::timestamptz
WHERE id = @id::uuid AND store_org_id = @store_org_id::uuid AND status = @previous_status::text;

-- name: ListNeedResponses :many
SELECT
    p.id, p.number, p.status, p.from_org_id, COALESCE(p.awaiting_org_id::text, '')::text AS awaiting_org_id,
    c.name AS consignor_name, p.created_at
FROM proposals AS p
JOIN needs AS n ON n.id = p.need_id
JOIN organizations AS c ON c.id = p.consignor_org_id
WHERE p.need_id = @need_id::uuid AND n.store_org_id = @store_org_id::uuid
ORDER BY p.created_at DESC, p.id DESC;

-- name: ExpireOpenNeeds :execrows
UPDATE needs
SET status = 'EXPIRED', updated_at = @now::timestamptz
WHERE status = 'OPEN' AND expires_at <= @now::timestamptz;

-- name: ListConsignorsNearNeed :many
SELECT o.id, ST_Distance(o.location_public, store.location, false)::float8 AS distance_meters
FROM organizations AS store
JOIN organizations AS o ON o.type = 'CONSIGNOR' AND o.status = 'ACTIVE' AND o.location_public IS NOT NULL
WHERE store.id = @store_org_id::uuid AND store.location IS NOT NULL
    AND ST_DWithin(o.location_public, store.location, @radius_meters::float8, false)
    AND EXISTS (
        SELECT 1 FROM products AS p
        WHERE p.consignor_org_id = o.id AND p.category_id = @category_id::uuid AND NOT p.is_archived
    )
ORDER BY distance_meters, o.id;

-- name: ClaimExpiringNeeds :many
UPDATE needs
SET reminded_at = @now::timestamptz
WHERE status = 'OPEN' AND reminded_at IS NULL AND expires_at > @now::timestamptz
    AND expires_at <= @remind_until::timestamptz
RETURNING id, store_org_id, title;
