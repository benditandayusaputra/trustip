-- name: LockConsignorForSKU :exec
SELECT id FROM organizations WHERE id = @id::uuid FOR UPDATE;

-- name: NextSKUNumber :one
SELECT (COALESCE(max(substring(sku FROM '[0-9]+$')::integer), 0) + 1)::integer
FROM products
WHERE consignor_org_id = @consignor_org_id::uuid AND sku ~ ('^' || @prefix::text || '-[0-9]{4,}$');

-- name: SaveProduct :execrows
INSERT INTO products (
    id, consignor_org_id, sku, name, category_id, unit, description, composition, net_content,
    photo_file_id, suggested_retail_price, suggested_deposit_price, shelf_life_days, created_at, updated_at
)
SELECT
    @id::uuid, @consignor_org_id::uuid, @sku::text, @name::text, c.id, @unit::text, @description::text,
    @composition::text, @net_content::text, NULLIF(@photo_file_id::text, '')::uuid,
    @suggested_retail_price::bigint, @suggested_deposit_price::bigint,
    sqlc.narg(shelf_life_days)::bigint, @changed_at::timestamptz, @changed_at::timestamptz
FROM categories AS c
WHERE c.id = @category_id::uuid
ON CONFLICT (id) DO UPDATE
SET sku = EXCLUDED.sku,
    name = EXCLUDED.name,
    category_id = EXCLUDED.category_id,
    unit = EXCLUDED.unit,
    description = EXCLUDED.description,
    composition = EXCLUDED.composition,
    net_content = EXCLUDED.net_content,
    photo_file_id = EXCLUDED.photo_file_id,
    suggested_retail_price = EXCLUDED.suggested_retail_price,
    suggested_deposit_price = EXCLUDED.suggested_deposit_price,
    shelf_life_days = EXCLUDED.shelf_life_days,
    updated_at = EXCLUDED.updated_at
WHERE products.consignor_org_id = EXCLUDED.consignor_org_id;

-- name: SetProductArchived :execrows
UPDATE products
SET is_archived = @is_archived::boolean, updated_at = @updated_at::timestamptz
WHERE id = @id::uuid AND consignor_org_id = @consignor_org_id::uuid;

-- name: GetProduct :one
SELECT
    p.id, p.consignor_org_id, p.sku, p.name, p.category_id, c.name AS category_name, p.unit, p.description,
    p.composition, p.net_content, COALESCE(p.photo_file_id::text, '')::text AS photo_file_id,
    COALESCE(f.storage_key, '')::text AS photo_storage_key, p.suggested_retail_price,
    p.suggested_deposit_price, p.shelf_life_days, p.is_archived, p.is_seeking_stores, o.slug AS consignor_slug,
    p.created_at, p.updated_at
FROM products AS p
JOIN categories AS c ON c.id = p.category_id
JOIN organizations AS o ON o.id = p.consignor_org_id
LEFT JOIN files AS f ON f.id = p.photo_file_id
WHERE p.id = @id::uuid AND p.consignor_org_id = @consignor_org_id::uuid;

-- name: ListProducts :many
SELECT
    p.id, p.consignor_org_id, p.sku, p.name, p.category_id, c.name AS category_name, p.unit, p.description,
    p.composition, p.net_content, COALESCE(p.photo_file_id::text, '')::text AS photo_file_id,
    COALESCE(f.storage_key, '')::text AS photo_storage_key, p.suggested_retail_price,
    p.suggested_deposit_price, p.shelf_life_days, p.is_archived, p.is_seeking_stores, o.slug AS consignor_slug,
    p.created_at, p.updated_at
FROM products AS p
JOIN categories AS c ON c.id = p.category_id
JOIN organizations AS o ON o.id = p.consignor_org_id
LEFT JOIN files AS f ON f.id = p.photo_file_id
WHERE p.consignor_org_id = @consignor_org_id::uuid
    AND p.is_archived = @is_archived::boolean
    AND (@category_id::text = '' OR p.category_id = NULLIF(@category_id::text, '')::uuid)
    AND (
        @query::text = ''
        OR p.name ILIKE '%' || replace(replace(replace(@query::text, '\', '\\'), '%', '\%'), '_', '\_') || '%'
        OR p.sku ILIKE '%' || replace(replace(replace(@query::text, '\', '\\'), '%', '\%'), '_', '\_') || '%'
    )
    AND (@from_start::boolean OR (p.created_at, p.id) < (@after_at::timestamptz, @after_id::uuid))
ORDER BY p.created_at DESC, p.id DESC
LIMIT @row_limit::bigint;

-- name: SetProductSeeking :execrows
UPDATE products
SET is_seeking_stores = @is_seeking_stores::boolean, updated_at = @updated_at::timestamptz
WHERE id = @id::uuid AND consignor_org_id = @consignor_org_id::uuid;
