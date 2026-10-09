-- name: GetPublicStoreProfile :one
SELECT
    o.id,
    o.slug,
    o.name,
    o.description,
    o.address,
    o.district,
    o.city_name,
    (o.location IS NOT NULL)::boolean AS has_location,
    COALESCE(ST_Y(o.location::geometry), 0)::float8 AS latitude,
    COALESCE(ST_X(o.location::geometry), 0)::float8 AS longitude,
    COALESCE(logo.storage_key, '')::text AS logo_storage_key,
    ARRAY(
        SELECT f.storage_key
        FROM org_photos AS ph
        JOIN files AS f ON f.id = ph.file_id
        WHERE ph.organization_id = o.id
        ORDER BY ph.sort_order
    )::text[] AS photo_storage_keys,
    COALESCE(sp.store_type, '')::text AS store_type,
    COALESCE(sp.opening_hours, '')::text AS opening_hours,
    COALESCE(sp.accepts_fixed_deposit, false)::boolean AS accepts_fixed_deposit,
    COALESCE(sp.accepts_commission, false)::boolean AS accepts_commission,
    (sp.commission_min_bps IS NOT NULL)::boolean AS has_commission_range,
    COALESCE(sp.commission_min_bps, 0)::bigint AS commission_min_bps,
    COALESCE(sp.commission_max_bps, 0)::bigint AS commission_max_bps,
    COALESCE(sp.requires_distribution_permit, false)::boolean AS requires_distribution_permit,
    COALESCE(sp.requires_halal, false)::boolean AS requires_halal,
    COALESCE(sp.requires_sample, false)::boolean AS requires_sample,
    COALESCE(sp.other_requirements, '')::text AS other_requirements,
    ARRAY(
        SELECT c.id::text
        FROM store_accepted_categories AS sac
        JOIN categories AS c ON c.id = sac.category_id
        WHERE sac.organization_id = o.id
        ORDER BY c.sort_order
    )::text[] AS accepted_category_ids,
    ARRAY(
        SELECT c.name
        FROM store_accepted_categories AS sac
        JOIN categories AS c ON c.id = sac.category_id
        WHERE sac.organization_id = o.id
        ORDER BY c.sort_order
    )::text[] AS accepted_category_names,
    COALESCE(st.settlements_counted, 0)::bigint AS due_settlements,
    COALESCE(st.on_time_payment_bps, 0)::bigint AS on_time_bps,
    COALESCE(st.active_partner_count, 0)::bigint AS active_partners,
    o.created_at
FROM organizations AS o
LEFT JOIN store_profiles AS sp ON sp.organization_id = o.id
LEFT JOIN org_stats AS st ON st.organization_id = o.id
LEFT JOIN files AS logo ON logo.id = o.logo_file_id
WHERE o.slug = @slug::text AND o.type = 'STORE' AND o.status = 'ACTIVE' AND o.is_listed;

-- name: ListPublicStoreNeeds :many
SELECT n.id, n.title, c.name AS category_name, n.expires_at
FROM needs AS n
JOIN categories AS c ON c.id = n.category_id
WHERE n.store_org_id = @store_org_id::uuid AND n.status = 'OPEN' AND n.expires_at > now()
ORDER BY n.created_at DESC, n.id DESC;

-- name: GetPublicConsignorProfile :one
SELECT
    o.id,
    o.slug,
    o.name,
    o.description,
    o.district,
    o.city_name,
    (o.location_public IS NOT NULL)::boolean AS has_location,
    COALESCE(ST_Y(o.location_public::geometry), 0)::float8 AS latitude,
    COALESCE(ST_X(o.location_public::geometry), 0)::float8 AS longitude,
    COALESCE(logo.storage_key, '')::text AS logo_storage_key,
    COALESCE(cp.business_id_no, '')::text AS business_id_no,
    COALESCE(cp.distribution_permit_no, '')::text AS distribution_permit_no,
    COALESCE(cp.halal_cert_no, '')::text AS halal_cert_no,
    COALESCE(st.active_partner_count, 0)::bigint AS active_partners,
    o.created_at
FROM organizations AS o
LEFT JOIN consignor_profiles AS cp ON cp.organization_id = o.id
LEFT JOIN org_stats AS st ON st.organization_id = o.id
LEFT JOIN files AS logo ON logo.id = o.logo_file_id
WHERE o.slug = @slug::text AND o.type = 'CONSIGNOR' AND o.status = 'ACTIVE' AND o.is_listed;

-- name: ListPublicConsignorProducts :many
SELECT
    p.sku,
    p.name,
    p.unit,
    c.name AS category_name,
    p.suggested_retail_price,
    COALESCE(f.storage_key, '')::text AS photo_storage_key
FROM products AS p
JOIN categories AS c ON c.id = p.category_id
LEFT JOIN files AS f ON f.id = p.photo_file_id
WHERE p.consignor_org_id = @consignor_org_id::uuid AND NOT p.is_archived
ORDER BY p.name, p.id
LIMIT @row_limit;

-- name: GetDirectoryVisitorCity :one
SELECT
    (city.code IS NOT NULL)::boolean AS has_city,
    COALESCE(ST_Y(city.centroid::geometry), 0)::float8 AS city_latitude,
    COALESCE(ST_X(city.centroid::geometry), 0)::float8 AS city_longitude
FROM (SELECT 1) AS anchor
LEFT JOIN cities AS city ON city.code = @city_code::text;

-- name: GetPublicNeed :one
SELECT
    n.id,
    n.title,
    c.name AS category_name,
    n.description,
    COALESCE(n.est_qty_per_week, 0)::bigint AS est_qty_per_week,
    COALESCE(n.max_retail_price, 0)::bigint AS max_retail_price,
    COALESCE(n.preferred_model, '')::text AS preferred_model,
    (n.offered_commission_bps IS NOT NULL)::boolean AS has_offered_commission,
    COALESCE(n.offered_commission_bps, 0)::bigint AS offered_commission_bps,
    (CASE WHEN n.status = 'OPEN' AND n.expires_at <= now() THEN 'EXPIRED' ELSE n.status END)::text AS status,
    n.expires_at,
    s.id AS store_id,
    s.name AS store_name,
    (CASE WHEN s.is_listed THEN s.slug ELSE '' END)::text AS store_profile_slug,
    COALESCE(sp.store_type, '')::text AS store_type,
    s.district AS store_district,
    s.city_name AS store_city_name
FROM needs AS n
JOIN categories AS c ON c.id = n.category_id
JOIN organizations AS s ON s.id = n.store_org_id
LEFT JOIN store_profiles AS sp ON sp.organization_id = s.id
WHERE n.id = @id::uuid AND s.status = 'ACTIVE';
