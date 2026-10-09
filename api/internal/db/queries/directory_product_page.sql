-- name: GetPublicProduct :one
SELECT
    p.id,
    p.sku,
    p.name,
    p.unit,
    p.description,
    p.composition,
    p.net_content,
    c.name AS category_name,
    p.suggested_retail_price,
    COALESCE(photo.storage_key, '')::text AS photo_storage_key,
    o.id AS consignor_id,
    (CASE WHEN o.status = 'ACTIVE' AND o.is_listed THEN o.slug ELSE '' END)::text AS consignor_profile_slug,
    o.name AS consignor_name,
    o.district AS consignor_district,
    o.city_name AS consignor_city_name,
    COALESCE(cp.business_id_no, '')::text AS business_id_no,
    COALESCE(cp.distribution_permit_no, '')::text AS distribution_permit_no,
    COALESCE(cp.halal_cert_no, '')::text AS halal_cert_no
FROM products AS p
JOIN organizations AS o ON o.id = p.consignor_org_id
JOIN categories AS c ON c.id = p.category_id
LEFT JOIN consignor_profiles AS cp ON cp.organization_id = o.id
LEFT JOIN files AS photo ON photo.id = p.photo_file_id
WHERE o.slug = @consignor_slug::text AND p.sku = @sku::text AND NOT p.is_archived;

-- name: ListPublicProductStores :many
WITH origin AS (
    SELECT ST_SetSRID(ST_MakePoint(@origin_longitude::float8, @origin_latitude::float8), 4326)::geography AS point
)
SELECT
    (CASE WHEN s.is_listed THEN s.slug ELSE '' END)::text AS profile_slug,
    s.name,
    s.address,
    s.district,
    s.city_name,
    (s.location IS NOT NULL)::boolean AS has_location,
    COALESCE(ST_Y(s.location::geometry), 0)::float8 AS latitude,
    COALESCE(ST_X(s.location::geometry), 0)::float8 AS longitude,
    (CASE
        WHEN @has_origin::boolean AND s.location IS NOT NULL THEN ST_Distance(s.location, origin.point)
        ELSE -1
    END)::float8 AS distance_meters
FROM partnerships AS pa
JOIN organizations AS s ON s.id = pa.store_org_id
CROSS JOIN origin
WHERE pa.consignor_org_id = @consignor_org_id::uuid
    AND pa.status = 'ACTIVE'
    AND pa.show_on_public
    AND s.status = 'ACTIVE'
    AND EXISTS (
        SELECT 1 FROM batch_balances AS b
        WHERE b.partnership_id = pa.id AND b.product_id = @product_id::uuid AND b.balance > 0
    )
ORDER BY
    (CASE WHEN @has_origin::boolean AND s.location IS NOT NULL THEN ST_Distance(s.location, origin.point) END)
        ASC NULLS LAST,
    s.name,
    s.id
LIMIT @row_limit;
