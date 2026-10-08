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
    (city.code IS NOT NULL)::boolean AS has_city,
    COALESCE(ST_Y(city.centroid::geometry), 0)::float8 AS city_latitude,
    COALESCE(ST_X(city.centroid::geometry), 0)::float8 AS city_longitude
FROM organizations AS o
LEFT JOIN consignor_profiles AS cp ON cp.organization_id = o.id
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
        COALESCE(o.last_active_at, '0001-01-01T00:00:00Z'::timestamptz)::timestamptz AS last_active_at,
        COALESCE(sp.store_type, '')::text AS store_type, sp.accepts_fixed_deposit, sp.accepts_commission,
        COALESCE(sp.commission_min_bps, -1)::bigint AS commission_min_bps,
        COALESCE(sp.commission_max_bps, -1)::bigint AS commission_max_bps,
        sp.requires_distribution_permit, sp.requires_halal, sp.requires_sample,
        COALESCE(f.storage_key, '')::text AS logo_storage_key,
        COALESCE(st.settlements_counted, 0)::bigint AS settlements_counted,
        COALESCE(st.on_time_payment_bps, 0)::bigint AS on_time_payment_bps,
        CASE WHEN @has_origin::boolean THEN ST_Distance(o.location_public, origin.point, false) ELSE -1 END::float8
            AS distance_meters,
        ARRAY(
            SELECT c.id::text FROM store_accepted_categories AS sac JOIN categories AS c ON c.id = sac.category_id
            WHERE sac.organization_id = o.id ORDER BY c.sort_order
        )::text[] AS accepted_category_ids,
        ARRAY(
            SELECT c.name FROM store_accepted_categories AS sac JOIN categories AS c ON c.id = sac.category_id
            WHERE sac.organization_id = o.id ORDER BY c.sort_order
        )::text[] AS accepted_category_names,
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
            @query::text = ''
            OR o.name ILIKE '%' || replace(replace(replace(@query::text, '\', '\\'), '%', '\%'), '_', '\_') || '%'
            OR o.description ILIKE '%' || replace(replace(replace(@query::text, '\', '\\'), '%', '\%'), '_', '\_')
                || '%'
        )
), ranked AS (
    SELECT
        listed.id, listed.slug, listed.name, listed.city_name, listed.district, listed.created_at,
        listed.last_active_at, listed.store_type, listed.accepts_fixed_deposit, listed.accepts_commission,
        listed.commission_min_bps, listed.commission_max_bps, listed.requires_distribution_permit,
        listed.requires_halal, listed.requires_sample, listed.logo_storage_key, listed.settlements_counted,
        listed.on_time_payment_bps, listed.distance_meters, listed.accepted_category_ids,
        listed.accepted_category_names, listed.is_self, listed.is_partner,
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
    requires_sample, logo_storage_key, settlements_counted, on_time_payment_bps, distance_meters,
    accepted_category_ids, accepted_category_names, is_self, is_partner, sort_text, sort_number
FROM ranked
WHERE @from_start::boolean OR (sort_text, sort_number, id) > (@after_text::text, @after_number::float8, @after_id::uuid)
ORDER BY sort_text, sort_number, id
LIMIT @row_limit::bigint;
