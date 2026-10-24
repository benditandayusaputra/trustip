-- name: CountDirectoryConsignorsByRadius :many
WITH origin AS (
    SELECT ST_SetSRID(ST_MakePoint(@origin_longitude::float8, @origin_latitude::float8), 4326)::geography AS point
), viewer AS (
    SELECT ST_SetSRID(ST_MakePoint(@viewer_longitude::float8, @viewer_latitude::float8), 4326)::geography AS point
), filtered AS (
    SELECT o.location_public
    FROM organizations AS o
    CROSS JOIN viewer
    JOIN consignor_profiles AS cp ON cp.organization_id = o.id
    WHERE o.type = 'CONSIGNOR' AND o.status = 'ACTIVE' AND o.is_listed
        AND (@viewer_id::uuid = '00000000-0000-0000-0000-000000000000'::uuid
            OR o.is_demo = EXISTS (SELECT 1 FROM organizations AS v WHERE v.id = @viewer_id::uuid AND v.is_demo))
        AND (NOT @has_permit::boolean OR cp.distribution_permit_no <> '')
        AND (NOT @has_halal::boolean OR cp.halal_cert_no <> '')
        AND (NOT @can_sample::boolean OR cp.can_provide_sample)
        AND (
            NOT @delivers_to_viewer::boolean
            OR ST_DWithin(o.location_public, viewer.point, cp.delivery_radius_km * 1000.0, false)
        )
        AND EXISTS (
            SELECT 1 FROM products AS p
            WHERE p.consignor_org_id = o.id AND NOT p.is_archived
                AND (NOT @seeking_only::boolean OR p.is_seeking_stores)
                AND (cardinality(@category_ids::uuid[]) = 0 OR p.category_id = ANY (@category_ids::uuid[]))
                AND (NOT @has_min_price::boolean OR p.suggested_retail_price >= @min_price::bigint)
                AND (NOT @has_max_price::boolean OR p.suggested_retail_price <= @max_price::bigint)
                AND (NOT @has_min_shelf_life::boolean OR p.shelf_life_days >= @min_shelf_life_days::bigint)
        )
        AND (
            @query::text = ''
            OR o.name ILIKE '%' || replace(replace(replace(@query::text, '\', '\\'), '%', '\%'), '_', '\_') || '%'
            OR o.description ILIKE '%' || replace(replace(replace(@query::text, '\', '\\'), '%', '\%'), '_', '\_')
                || '%'
            OR EXISTS (
                SELECT 1 FROM products AS p
                WHERE p.consignor_org_id = o.id AND NOT p.is_archived
                    AND p.name ILIKE '%' || replace(replace(replace(@query::text, '\', '\\'), '%', '\%'), '_', '\_')
                        || '%'
            )
        )
)
SELECT
    radius.meters::float8 AS radius_meters,
    count(filtered.location_public) FILTER (
        WHERE ST_DWithin(filtered.location_public, origin.point, radius.meters, false)
    )::bigint AS consignor_count
FROM unnest(@radii_meters::float8[]) AS radius (meters)
CROSS JOIN origin
LEFT JOIN filtered ON true
GROUP BY radius.meters
ORDER BY radius.meters;

-- name: ListDirectoryConsignors :many
WITH origin AS (
    SELECT ST_SetSRID(ST_MakePoint(@origin_longitude::float8, @origin_latitude::float8), 4326)::geography AS point
), viewer AS (
    SELECT ST_SetSRID(ST_MakePoint(@viewer_longitude::float8, @viewer_latitude::float8), 4326)::geography AS point
), listed AS (
    SELECT
        o.id, o.slug, o.name, o.city_name, o.district, o.created_at,
        COALESCE(o.last_active_at, '0001-01-01T00:00:00Z'::timestamptz)::timestamptz AS last_active_at,
        COALESCE(f.storage_key, '')::text AS logo_storage_key,
        COALESCE(cp.delivery_radius_km, 0)::bigint AS delivery_radius_km,
        cp.preferred_store_types, (cp.distribution_permit_no <> '')::boolean AS has_permit,
        (cp.halal_cert_no <> '')::boolean AS has_halal, cp.can_provide_sample,
        COALESCE(st.active_partner_count, 0)::bigint AS active_partner_count,
        CASE WHEN @has_origin::boolean THEN ST_Distance(o.location_public, origin.point, false) ELSE -1 END::float8
            AS distance_meters,
        ARRAY(
            SELECT c.id::text FROM categories AS c
            WHERE EXISTS (
                SELECT 1 FROM products AS p
                WHERE p.consignor_org_id = o.id AND p.category_id = c.id AND NOT p.is_archived
            )
            ORDER BY c.sort_order
        )::text[] AS category_ids,
        ARRAY(
            SELECT c.name FROM categories AS c
            WHERE EXISTS (
                SELECT 1 FROM products AS p
                WHERE p.consignor_org_id = o.id AND p.category_id = c.id AND NOT p.is_archived
            )
            ORDER BY c.sort_order
        )::text[] AS category_names,
        (
            SELECT count(*) FROM products AS p WHERE p.consignor_org_id = o.id AND NOT p.is_archived
        )::bigint AS product_count,
        (o.id = @viewer_id::uuid)::boolean AS is_self,
        EXISTS (
            SELECT 1 FROM partnerships AS pa
            WHERE pa.consignor_org_id = o.id AND pa.store_org_id = @viewer_id::uuid AND pa.status = 'ACTIVE'
        )::boolean AS is_partner
    FROM organizations AS o
    CROSS JOIN origin
    CROSS JOIN viewer
    JOIN consignor_profiles AS cp ON cp.organization_id = o.id
    LEFT JOIN org_stats AS st ON st.organization_id = o.id
    LEFT JOIN files AS f ON f.id = o.logo_file_id
    WHERE o.type = 'CONSIGNOR' AND o.status = 'ACTIVE' AND o.is_listed
        AND (@viewer_id::uuid = '00000000-0000-0000-0000-000000000000'::uuid
            OR o.is_demo = EXISTS (SELECT 1 FROM organizations AS v WHERE v.id = @viewer_id::uuid AND v.is_demo))
        AND (
            NOT @within_radius::boolean
            OR ST_DWithin(o.location_public, origin.point, @radius_meters::float8, false)
        )
        AND (NOT @has_permit::boolean OR cp.distribution_permit_no <> '')
        AND (NOT @has_halal::boolean OR cp.halal_cert_no <> '')
        AND (NOT @can_sample::boolean OR cp.can_provide_sample)
        AND (
            NOT @delivers_to_viewer::boolean
            OR ST_DWithin(o.location_public, viewer.point, cp.delivery_radius_km * 1000.0, false)
        )
        AND EXISTS (
            SELECT 1 FROM products AS p
            WHERE p.consignor_org_id = o.id AND NOT p.is_archived
                AND (cardinality(@category_ids::uuid[]) = 0 OR p.category_id = ANY (@category_ids::uuid[]))
                AND (NOT @has_min_price::boolean OR p.suggested_retail_price >= @min_price::bigint)
                AND (NOT @has_max_price::boolean OR p.suggested_retail_price <= @max_price::bigint)
                AND (NOT @has_min_shelf_life::boolean OR p.shelf_life_days >= @min_shelf_life_days::bigint)
        )
        AND (
            @query::text = ''
            OR o.name ILIKE '%' || replace(replace(replace(@query::text, '\', '\\'), '%', '\%'), '_', '\_') || '%'
            OR o.description ILIKE '%' || replace(replace(replace(@query::text, '\', '\\'), '%', '\%'), '_', '\_')
                || '%'
            OR EXISTS (
                SELECT 1 FROM products AS p
                WHERE p.consignor_org_id = o.id AND NOT p.is_archived
                    AND p.name ILIKE '%' || replace(replace(replace(@query::text, '\', '\\'), '%', '\%'), '_', '\_')
                        || '%'
            )
        )
), ranked AS (
    SELECT
        listed.id, listed.slug, listed.name, listed.city_name, listed.district, listed.created_at,
        listed.last_active_at, listed.logo_storage_key, listed.delivery_radius_km, listed.preferred_store_types,
        listed.has_permit, listed.has_halal, listed.can_provide_sample, listed.active_partner_count,
        listed.distance_meters, listed.category_ids, listed.category_names, listed.product_count, listed.is_self,
        listed.is_partner,
        CASE @sort_order::text WHEN 'CITY' THEN listed.city_name ELSE '' END::text AS sort_text,
        CASE @sort_order::text
            WHEN 'NEAREST' THEN listed.distance_meters
            WHEN 'REPUTATION' THEN -listed.active_partner_count
            ELSE -extract(epoch FROM listed.created_at)
        END::float8 AS sort_number
    FROM listed
)
SELECT
    id, slug, name, city_name, district, created_at, last_active_at, logo_storage_key, delivery_radius_km,
    preferred_store_types, has_permit, has_halal, can_provide_sample, active_partner_count, distance_meters,
    category_ids, category_names, product_count, is_self, is_partner, sort_text, sort_number
FROM ranked
WHERE @from_start::boolean OR (sort_text, sort_number, id) > (@after_text::text, @after_number::float8, @after_id::uuid)
ORDER BY sort_text, sort_number, id
LIMIT @row_limit::bigint;

-- name: ListDirectoryProducts :many
WITH origin AS (
    SELECT ST_SetSRID(ST_MakePoint(@origin_longitude::float8, @origin_latitude::float8), 4326)::geography AS point
), viewer AS (
    SELECT ST_SetSRID(ST_MakePoint(@viewer_longitude::float8, @viewer_latitude::float8), 4326)::geography AS point
), listed AS (
    SELECT
        p.id, p.name, p.unit, p.category_id::text AS category_id, c.name AS category_name,
        p.suggested_retail_price, COALESCE(p.shelf_life_days, 0)::bigint AS shelf_life_days,
        COALESCE(pf.storage_key, '')::text AS photo_storage_key, p.created_at, p.sku,
        o.id AS consignor_id, o.slug AS consignor_slug, o.name AS consignor_name, o.city_name,
        COALESCE(o.last_active_at, '0001-01-01T00:00:00Z'::timestamptz)::timestamptz AS last_active_at,
        COALESCE(cp.delivery_radius_km, 0)::bigint AS delivery_radius_km,
        cp.preferred_store_types, (cp.distribution_permit_no <> '')::boolean AS has_permit,
        (cp.halal_cert_no <> '')::boolean AS has_halal, cp.can_provide_sample,
        COALESCE(st.active_partner_count, 0)::bigint AS active_partner_count,
        CASE WHEN @has_origin::boolean THEN ST_Distance(o.location_public, origin.point, false) ELSE -1 END::float8
            AS distance_meters,
        (o.id = @viewer_id::uuid)::boolean AS is_self,
        EXISTS (
            SELECT 1 FROM partnerships AS pa
            WHERE pa.consignor_org_id = o.id AND pa.store_org_id = @viewer_id::uuid AND pa.status = 'ACTIVE'
        )::boolean AS is_partner
    FROM products AS p
    JOIN categories AS c ON c.id = p.category_id
    JOIN organizations AS o ON o.id = p.consignor_org_id
    CROSS JOIN origin
    CROSS JOIN viewer
    JOIN consignor_profiles AS cp ON cp.organization_id = o.id
    LEFT JOIN org_stats AS st ON st.organization_id = o.id
    LEFT JOIN files AS pf ON pf.id = p.photo_file_id
    WHERE p.is_seeking_stores AND NOT p.is_archived
        AND o.type = 'CONSIGNOR' AND o.status = 'ACTIVE' AND o.is_listed
        AND (@viewer_id::uuid = '00000000-0000-0000-0000-000000000000'::uuid
            OR o.is_demo = EXISTS (SELECT 1 FROM organizations AS v WHERE v.id = @viewer_id::uuid AND v.is_demo))
        AND (@consignor_id::text = '' OR o.id = NULLIF(@consignor_id::text, '')::uuid)
        AND (
            NOT @within_radius::boolean
            OR ST_DWithin(o.location_public, origin.point, @radius_meters::float8, false)
        )
        AND (cardinality(@category_ids::uuid[]) = 0 OR p.category_id = ANY (@category_ids::uuid[]))
        AND (NOT @has_min_price::boolean OR p.suggested_retail_price >= @min_price::bigint)
        AND (NOT @has_max_price::boolean OR p.suggested_retail_price <= @max_price::bigint)
        AND (NOT @has_min_shelf_life::boolean OR p.shelf_life_days >= @min_shelf_life_days::bigint)
        AND (NOT @has_permit::boolean OR cp.distribution_permit_no <> '')
        AND (NOT @has_halal::boolean OR cp.halal_cert_no <> '')
        AND (NOT @can_sample::boolean OR cp.can_provide_sample)
        AND (
            NOT @delivers_to_viewer::boolean
            OR ST_DWithin(o.location_public, viewer.point, cp.delivery_radius_km * 1000.0, false)
        )
        AND (
            @query::text = ''
            OR p.name ILIKE '%' || replace(replace(replace(@query::text, '\', '\\'), '%', '\%'), '_', '\_') || '%'
            OR p.description ILIKE '%' || replace(replace(replace(@query::text, '\', '\\'), '%', '\%'), '_', '\_')
                || '%'
            OR o.name ILIKE '%' || replace(replace(replace(@query::text, '\', '\\'), '%', '\%'), '_', '\_') || '%'
        )
), ranked AS (
    SELECT
        listed.id, listed.name, listed.unit, listed.category_id, listed.category_name,
        listed.suggested_retail_price, listed.shelf_life_days, listed.photo_storage_key, listed.created_at,
        listed.sku, listed.consignor_id, listed.consignor_slug, listed.consignor_name, listed.city_name,
        listed.last_active_at, listed.delivery_radius_km, listed.preferred_store_types, listed.has_permit,
        listed.has_halal, listed.can_provide_sample, listed.active_partner_count, listed.distance_meters,
        listed.is_self, listed.is_partner,
        CASE @sort_order::text WHEN 'CITY' THEN listed.city_name ELSE '' END::text AS sort_text,
        CASE @sort_order::text
            WHEN 'NEAREST' THEN listed.distance_meters
            WHEN 'REPUTATION' THEN -listed.active_partner_count
            ELSE -extract(epoch FROM listed.created_at)
        END::float8 AS sort_number
    FROM listed
)
SELECT
    id, name, unit, category_id, category_name, suggested_retail_price, shelf_life_days, photo_storage_key,
    created_at, sku, consignor_id, consignor_slug, consignor_name, city_name, last_active_at, delivery_radius_km,
    preferred_store_types, has_permit, has_halal, can_provide_sample, active_partner_count, distance_meters,
    is_self, is_partner, sort_text, sort_number
FROM ranked
WHERE @from_start::boolean OR (sort_text, sort_number, id) > (@after_text::text, @after_number::float8, @after_id::uuid)
ORDER BY sort_text, sort_number, id
LIMIT @row_limit::bigint;
