-- name: GetDirectoryViewer :one
SELECT
    o.type,
    (o.location IS NOT NULL)::boolean AS has_location,
    COALESCE(ST_Y(o.location::geometry), 0)::float8 AS latitude,
    COALESCE(ST_X(o.location::geometry), 0)::float8 AS longitude,
    COALESCE(cp.delivery_radius_km, 0)::bigint AS delivery_radius_km,
    COALESCE(cp.preferred_store_types, '{}')::text[] AS preferred_store_types,
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
    COALESCE(sp.store_type, '')::text AS store_type,
    ARRAY(
        SELECT sac.category_id::text FROM store_accepted_categories AS sac WHERE sac.organization_id = o.id
    )::text[] AS accepted_category_ids,
    ARRAY(
        SELECT DISTINCT n.category_id::text FROM needs AS n
        WHERE n.store_org_id = o.id AND n.status = 'OPEN' AND n.expires_at > @now::timestamptz
    )::text[] AS need_category_ids,
    (city.code IS NOT NULL)::boolean AS has_city,
    COALESCE(ST_Y(city.centroid::geometry), 0)::float8 AS city_latitude,
    COALESCE(ST_X(city.centroid::geometry), 0)::float8 AS city_longitude
FROM organizations AS o
LEFT JOIN consignor_profiles AS cp ON cp.organization_id = o.id
LEFT JOIN store_profiles AS sp ON sp.organization_id = o.id
LEFT JOIN cities AS city ON city.code = @city_code::text
WHERE o.id = @organization_id::uuid;

-- name: CountDirectoryStoresByRadius :many
WITH origin AS (
    SELECT ST_SetSRID(ST_MakePoint(@origin_longitude::float8, @origin_latitude::float8), 4326)::geography AS point
), filtered AS (
    SELECT o.location_public
    FROM organizations AS o
    JOIN store_profiles AS sp ON sp.organization_id = o.id
    WHERE o.type = 'STORE' AND o.status = 'ACTIVE' AND o.is_listed
        AND o.is_demo = EXISTS (SELECT 1 FROM organizations AS v WHERE v.id = NULLIF(@viewer_id::text, '')::uuid AND v.is_demo)
        AND (@store_type::text = '' OR sp.store_type = @store_type::text)
        AND (
            cardinality(@category_ids::uuid[]) = 0
            OR EXISTS (
                SELECT 1 FROM store_accepted_categories AS sac
                WHERE sac.organization_id = o.id AND sac.category_id = ANY (@category_ids::uuid[])
            )
        )
        AND (
            @pricing_model::text = ''
            OR (@pricing_model::text = 'FIXED_DEPOSIT' AND sp.accepts_fixed_deposit)
            OR (@pricing_model::text = 'COMMISSION' AND sp.accepts_commission)
        )
        AND (NOT @has_max_commission::boolean OR sp.commission_min_bps <= @max_commission_bps::bigint)
        AND (NOT @without_permit::boolean OR NOT sp.requires_distribution_permit)
        AND (NOT @without_halal::boolean OR NOT sp.requires_halal)
        AND (
            NOT @has_open_need::boolean
            OR EXISTS (
                SELECT 1 FROM needs AS n
                WHERE n.store_org_id = o.id AND n.status = 'OPEN' AND n.expires_at > @now::timestamptz
            )
        )
        AND (
            @query::text = ''
            OR o.name ILIKE '%' || replace(replace(replace(@query::text, '\', '\\'), '%', '\%'), '_', '\_') || '%'
            OR o.description ILIKE '%' || replace(replace(replace(@query::text, '\', '\\'), '%', '\%'), '_', '\_')
                || '%'
        )
)
SELECT
    radius.meters::float8 AS radius_meters,
    count(filtered.location_public) FILTER (
        WHERE ST_DWithin(filtered.location_public, origin.point, radius.meters, false)
    )::bigint AS store_count
FROM unnest(@radii_meters::float8[]) AS radius (meters)
CROSS JOIN origin
LEFT JOIN filtered ON true
GROUP BY radius.meters
ORDER BY radius.meters;

-- name: ListDirectoryStores :many
WITH origin AS (
    SELECT ST_SetSRID(ST_MakePoint(@origin_longitude::float8, @origin_latitude::float8), 4326)::geography AS point
), listed AS (
    SELECT
        o.id, o.slug, o.name, o.city_name, o.district, o.created_at,
        (o.location_public IS NOT NULL)::boolean AS has_location,
        COALESCE(ST_Y(o.location_public::geometry), 0)::float8 AS latitude,
        COALESCE(ST_X(o.location_public::geometry), 0)::float8 AS longitude,
        COALESCE(o.last_active_at, '0001-01-01T00:00:00Z'::timestamptz)::timestamptz AS last_active_at,
        COALESCE(sp.store_type, '')::text AS store_type, sp.accepts_fixed_deposit, sp.accepts_commission,
        COALESCE(sp.commission_min_bps, -1)::bigint AS commission_min_bps,
        COALESCE(sp.commission_max_bps, -1)::bigint AS commission_max_bps,
        sp.requires_distribution_permit, sp.requires_halal, sp.requires_sample,
        COALESCE(f.storage_key, '')::text AS logo_storage_key,
        COALESCE(st.settlements_counted, 0)::bigint AS settlements_counted,
        COALESCE(st.on_time_payment_bps, 0)::bigint AS on_time_payment_bps,
        COALESCE(st.active_partner_count, 0)::bigint AS active_partner_count,
        CASE WHEN @has_origin::boolean THEN COALESCE(ST_Distance(o.location_public, origin.point, false), -1) ELSE -1 END::float8
            AS distance_meters,
        ARRAY(
            SELECT c.id::text FROM store_accepted_categories AS sac JOIN categories AS c ON c.id = sac.category_id
            WHERE sac.organization_id = o.id ORDER BY c.sort_order
        )::text[] AS accepted_category_ids,
        ARRAY(
            SELECT c.name FROM store_accepted_categories AS sac JOIN categories AS c ON c.id = sac.category_id
            WHERE sac.organization_id = o.id ORDER BY c.sort_order
        )::text[] AS accepted_category_names,
        ARRAY(
            SELECT DISTINCT n.category_id::text FROM needs AS n
            WHERE n.store_org_id = o.id AND n.status = 'OPEN' AND n.expires_at > @now::timestamptz
        )::text[] AS need_category_ids,
        (o.id = @viewer_id::uuid)::boolean AS is_self,
        EXISTS (
            SELECT 1 FROM partnerships AS pa
            WHERE pa.store_org_id = o.id AND pa.consignor_org_id = @viewer_id::uuid AND pa.status = 'ACTIVE'
        )::boolean AS is_partner
    FROM organizations AS o
    CROSS JOIN origin
    JOIN store_profiles AS sp ON sp.organization_id = o.id
    LEFT JOIN org_stats AS st ON st.organization_id = o.id
    LEFT JOIN files AS f ON f.id = o.logo_file_id
    WHERE o.type = 'STORE' AND o.status = 'ACTIVE' AND o.is_listed
        AND o.is_demo = EXISTS (SELECT 1 FROM organizations AS v WHERE v.id = @viewer_id::uuid AND v.is_demo)
        AND (
            NOT @within_radius::boolean
            OR ST_DWithin(o.location_public, origin.point, @radius_meters::float8, false)
        )
        AND (@store_type::text = '' OR sp.store_type = @store_type::text)
        AND (
            cardinality(@category_ids::uuid[]) = 0
            OR EXISTS (
                SELECT 1 FROM store_accepted_categories AS sac
                WHERE sac.organization_id = o.id AND sac.category_id = ANY (@category_ids::uuid[])
            )
        )
        AND (
            @pricing_model::text = ''
            OR (@pricing_model::text = 'FIXED_DEPOSIT' AND sp.accepts_fixed_deposit)
            OR (@pricing_model::text = 'COMMISSION' AND sp.accepts_commission)
        )
        AND (NOT @has_max_commission::boolean OR sp.commission_min_bps <= @max_commission_bps::bigint)
        AND (NOT @without_permit::boolean OR NOT sp.requires_distribution_permit)
        AND (NOT @without_halal::boolean OR NOT sp.requires_halal)
        AND (
            NOT @has_open_need::boolean
            OR EXISTS (
                SELECT 1 FROM needs AS n
                WHERE n.store_org_id = o.id AND n.status = 'OPEN' AND n.expires_at > @now::timestamptz
            )
        )
        AND (
            @query::text = ''
            OR o.name ILIKE '%' || replace(replace(replace(@query::text, '\', '\\'), '%', '\%'), '_', '\_') || '%'
            OR o.description ILIKE '%' || replace(replace(replace(@query::text, '\', '\\'), '%', '\%'), '_', '\_')
                || '%'
        )
), ranked AS (
    SELECT
        listed.id, listed.slug, listed.name, listed.city_name, listed.district, listed.created_at,
        listed.has_location, listed.latitude, listed.longitude,
        listed.last_active_at, listed.store_type, listed.accepts_fixed_deposit, listed.accepts_commission,
        listed.commission_min_bps, listed.commission_max_bps, listed.requires_distribution_permit,
        listed.requires_halal, listed.requires_sample, listed.logo_storage_key, listed.settlements_counted,
        listed.on_time_payment_bps, listed.active_partner_count, listed.distance_meters,
        listed.accepted_category_ids, listed.accepted_category_names, listed.need_category_ids, listed.is_self,
        listed.is_partner,
        CASE @sort_order::text WHEN 'CITY' THEN listed.city_name ELSE '' END::text AS sort_text,
        CASE @sort_order::text
            WHEN 'NEAREST' THEN listed.distance_meters
            WHEN 'REPUTATION' THEN -CASE
                WHEN listed.settlements_counted >= @rated_from::bigint THEN listed.on_time_payment_bps ELSE -1
            END
            ELSE -extract(epoch FROM listed.created_at)
        END::float8 AS sort_number
    FROM listed
)
SELECT
    id, slug, name, city_name, district, created_at, last_active_at, store_type, accepts_fixed_deposit,
    accepts_commission, commission_min_bps, commission_max_bps, requires_distribution_permit, requires_halal,
    requires_sample, logo_storage_key, settlements_counted, on_time_payment_bps, active_partner_count,
    distance_meters, accepted_category_ids, accepted_category_names, need_category_ids, is_self, is_partner,
    sort_text, sort_number, has_location, latitude, longitude
FROM ranked
WHERE @from_start::boolean OR (sort_text, sort_number, id) > (@after_text::text, @after_number::float8, @after_id::uuid)
ORDER BY sort_text, sort_number, id
LIMIT @row_limit::bigint;

-- name: GetDirectoryContact :one
SELECT o.id, o.name, o.type, o.city_name, o.whatsapp
FROM organizations AS o
WHERE o.id = @id::uuid AND o.status = 'ACTIVE' AND o.is_listed;

-- name: CountDirectoryNeedsByRadius :many
WITH origin AS (
    SELECT ST_SetSRID(ST_MakePoint(@origin_longitude::float8, @origin_latitude::float8), 4326)::geography AS point
), filtered AS (
    SELECT o.location_public
    FROM needs AS n
    JOIN organizations AS o ON o.id = n.store_org_id
    WHERE n.status = 'OPEN' AND n.expires_at > @now::timestamptz AND o.status = 'ACTIVE' AND o.is_listed
        AND o.is_demo = EXISTS (SELECT 1 FROM organizations AS v WHERE v.id = @viewer_id::uuid AND v.is_demo)
        AND (cardinality(@category_ids::uuid[]) = 0 OR n.category_id = ANY (@category_ids::uuid[]))
        AND (
            @query::text = ''
            OR n.title ILIKE '%' || replace(replace(replace(@query::text, '\', '\\'), '%', '\%'), '_', '\_') || '%'
            OR n.description ILIKE '%' || replace(replace(replace(@query::text, '\', '\\'), '%', '\%'), '_', '\_')
                || '%'
        )
)
SELECT
    radius.meters::float8 AS radius_meters,
    count(filtered.location_public) FILTER (
        WHERE ST_DWithin(filtered.location_public, origin.point, radius.meters, false)
    )::bigint AS need_count
FROM unnest(@radii_meters::float8[]) AS radius (meters)
CROSS JOIN origin
LEFT JOIN filtered ON true
GROUP BY radius.meters
ORDER BY radius.meters;

-- name: ListDirectoryNeeds :many
WITH origin AS (
    SELECT ST_SetSRID(ST_MakePoint(@origin_longitude::float8, @origin_latitude::float8), 4326)::geography AS point
), listed AS (
    SELECT
        n.id, n.title, n.description, n.category_id::text AS category_id, c.name AS category_name,
        COALESCE(n.est_qty_per_week, 0)::bigint AS est_qty_per_week,
        COALESCE(n.max_retail_price, 0)::bigint AS max_retail_price,
        COALESCE(n.preferred_model, '')::text AS preferred_model,
        COALESCE(n.offered_commission_bps, -1)::bigint AS offered_commission_bps,
        n.expires_at, n.created_at, o.id AS store_id, o.slug AS store_slug, o.name AS store_name, o.city_name,
        COALESCE(sp.store_type, '')::text AS store_type,
        CASE WHEN @has_origin::boolean THEN ST_Distance(o.location_public, origin.point, false) ELSE -1 END::float8
            AS distance_meters,
        (o.id = @viewer_id::uuid)::boolean AS is_own
    FROM needs AS n
    JOIN categories AS c ON c.id = n.category_id
    JOIN organizations AS o ON o.id = n.store_org_id
    CROSS JOIN origin
    LEFT JOIN store_profiles AS sp ON sp.organization_id = o.id
    WHERE n.status = 'OPEN' AND n.expires_at > @now::timestamptz AND o.status = 'ACTIVE' AND o.is_listed
        AND o.is_demo = EXISTS (SELECT 1 FROM organizations AS v WHERE v.id = @viewer_id::uuid AND v.is_demo)
        AND (
            NOT @within_radius::boolean
            OR ST_DWithin(o.location_public, origin.point, @radius_meters::float8, false)
        )
        AND (cardinality(@category_ids::uuid[]) = 0 OR n.category_id = ANY (@category_ids::uuid[]))
        AND (
            @query::text = ''
            OR n.title ILIKE '%' || replace(replace(replace(@query::text, '\', '\\'), '%', '\%'), '_', '\_') || '%'
            OR n.description ILIKE '%' || replace(replace(replace(@query::text, '\', '\\'), '%', '\%'), '_', '\_')
                || '%'
        )
), ranked AS (
    SELECT
        listed.id, listed.title, listed.description, listed.category_id, listed.category_name,
        listed.est_qty_per_week, listed.max_retail_price, listed.preferred_model, listed.offered_commission_bps,
        listed.expires_at, listed.created_at, listed.store_id, listed.store_slug, listed.store_name,
        listed.city_name, listed.store_type, listed.distance_meters, listed.is_own,
        CASE @sort_order::text WHEN 'CITY' THEN listed.city_name ELSE '' END::text AS sort_text,
        CASE @sort_order::text
            WHEN 'NEAREST' THEN listed.distance_meters
            ELSE -extract(epoch FROM listed.created_at)
        END::float8 AS sort_number
    FROM listed
)
SELECT
    id, title, description, category_id, category_name, est_qty_per_week, max_retail_price, preferred_model,
    offered_commission_bps, expires_at, created_at, store_id, store_slug, store_name, city_name, store_type,
    distance_meters, is_own, sort_text, sort_number
FROM ranked
WHERE @from_start::boolean OR (sort_text, sort_number, id) > (@after_text::text, @after_number::float8, @after_id::uuid)
ORDER BY sort_text, sort_number, id
LIMIT @row_limit::bigint;
