-- name: CreateOrganization :one
INSERT INTO organizations (
    id, type, name, slug, description, whatsapp, city_code, city_name, district, address,
    location, location_public, created_at, updated_at
)
SELECT
    @id::uuid, @type::text, @name::text, @slug::text, @description::text, @whatsapp::text,
    c.code, c.name, @district::text, @address::text,
    CASE WHEN @has_location::boolean
        THEN ST_SetSRID(ST_MakePoint(@longitude::float8, @latitude::float8), 4326)::geography
    END,
    CASE WHEN @has_location::boolean
        THEN ST_SetSRID(ST_MakePoint(@public_longitude::float8, @public_latitude::float8), 4326)::geography
    END,
    @created_at::timestamptz, @created_at::timestamptz
FROM cities AS c
WHERE c.code = @city_code::text
RETURNING city_name;

-- name: CreateStoreProfile :exec
INSERT INTO store_profiles (organization_id, created_at, updated_at)
VALUES (@organization_id, @created_at, @created_at);

-- name: CreateConsignorProfile :exec
INSERT INTO consignor_profiles (organization_id, created_at, updated_at)
VALUES (@organization_id, @created_at, @created_at);

-- name: CreateMembership :exec
INSERT INTO memberships (id, organization_id, user_id, role, created_at, updated_at)
VALUES (@id, @organization_id, @user_id, @role, @created_at, @created_at);

-- name: ListTakenSlugs :many
SELECT slug
FROM organizations
WHERE slug ~ ('^' || @base::text || '(-[0-9]+)?$');

-- name: FindActiveMembership :one
SELECT m.organization_id, o.type AS organization_type, m.role
FROM memberships AS m
JOIN organizations AS o ON o.id = m.organization_id
WHERE m.user_id = @user_id AND m.is_active;

-- name: ListProvinces :many
SELECT DISTINCT left(code, 2)::text AS province_code, province_name
FROM cities
ORDER BY province_code;

-- name: ListCitiesInProvince :many
SELECT
    code, name, province_name,
    ST_Y(centroid::geometry)::float8 AS latitude,
    ST_X(centroid::geometry)::float8 AS longitude
FROM cities
WHERE left(code, 2) = @province_code::text
ORDER BY name;

-- name: GetOrganizationSettings :one
SELECT
    o.id, o.type, o.name, o.slug, o.description, o.whatsapp,
    COALESCE(o.city_code, '')::text AS city_code, o.city_name, o.district, o.address,
    (o.location IS NOT NULL)::boolean AS has_location,
    COALESCE(ST_Y(o.location::geometry), 0)::float8 AS latitude,
    COALESCE(ST_X(o.location::geometry), 0)::float8 AS longitude,
    COALESCE(ST_Y(o.location_public::geometry), 0)::float8 AS public_latitude,
    COALESCE(ST_X(o.location_public::geometry), 0)::float8 AS public_longitude,
    COALESCE(f.id::text, '')::text AS logo_file_id,
    COALESCE(f.storage_key, '')::text AS logo_storage_key
FROM organizations AS o
LEFT JOIN files AS f ON f.id = o.logo_file_id
WHERE o.id = @id;

-- name: CountOrganizationFiles :one
SELECT count(*)::integer
FROM files
WHERE id = @id::uuid AND organization_id = @organization_id AND purpose = @purpose;

-- name: UpdateOrganizationProfile :execrows
UPDATE organizations AS o
SET
    name = @name::text,
    description = @description::text,
    whatsapp = @whatsapp::text,
    city_code = c.code,
    city_name = c.name,
    district = @district::text,
    address = @address::text,
    location = CASE WHEN @has_location::boolean
        THEN ST_SetSRID(ST_MakePoint(@longitude::float8, @latitude::float8), 4326)::geography
    END,
    location_public = CASE WHEN @has_location::boolean
        THEN ST_SetSRID(ST_MakePoint(@public_longitude::float8, @public_latitude::float8), 4326)::geography
    END,
    logo_file_id = NULLIF(@logo_file_id::text, '')::uuid,
    updated_at = @updated_at::timestamptz
FROM cities AS c
WHERE o.id = @id::uuid AND c.code = @city_code::text;

-- name: ListCategories :many
SELECT id, name, slug
FROM categories
ORDER BY sort_order;

-- name: GetStoreProfile :one
SELECT
    COALESCE(sp.store_type, '')::text AS store_type,
    sp.opening_hours, sp.accepts_fixed_deposit, sp.accepts_commission,
    (sp.commission_min_bps IS NOT NULL)::boolean AS has_commission,
    COALESCE(sp.commission_min_bps, 0)::bigint AS commission_min_bps,
    COALESCE(sp.commission_max_bps, 0)::bigint AS commission_max_bps,
    sp.requires_distribution_permit, sp.requires_halal, sp.requires_sample, sp.other_requirements,
    COALESCE((
        SELECT array_agg(sac.category_id::text ORDER BY c.sort_order)
        FROM store_accepted_categories AS sac
        JOIN categories AS c ON c.id = sac.category_id
        WHERE sac.organization_id = sp.organization_id
    ), '{}')::text[] AS category_ids
FROM store_profiles AS sp
WHERE sp.organization_id = @organization_id;

-- name: ListOrganizationPhotos :many
SELECT f.id, f.storage_key
FROM org_photos AS p
JOIN files AS f ON f.id = p.file_id
WHERE p.organization_id = @organization_id
ORDER BY p.sort_order;

-- name: UpdateStoreProfile :execrows
UPDATE store_profiles
SET
    store_type = NULLIF(@store_type::text, ''),
    opening_hours = @opening_hours::text,
    accepts_fixed_deposit = @accepts_fixed_deposit::boolean,
    accepts_commission = @accepts_commission::boolean,
    commission_min_bps = CASE WHEN @has_commission::boolean THEN @commission_min_bps::bigint END,
    commission_max_bps = CASE WHEN @has_commission::boolean THEN @commission_max_bps::bigint END,
    requires_distribution_permit = @requires_distribution_permit::boolean,
    requires_halal = @requires_halal::boolean,
    requires_sample = @requires_sample::boolean,
    other_requirements = @other_requirements::text,
    updated_at = @updated_at::timestamptz
WHERE organization_id = @organization_id::uuid;

-- name: ClearAcceptedCategories :exec
DELETE FROM store_accepted_categories
WHERE organization_id = @organization_id;

-- name: AddAcceptedCategories :execrows
INSERT INTO store_accepted_categories (organization_id, category_id)
SELECT @organization_id::uuid, c.id
FROM categories AS c
WHERE c.id = ANY(@category_ids::uuid[]);

-- name: ClearOrganizationPhotos :exec
DELETE FROM org_photos
WHERE organization_id = @organization_id;

-- name: AddOrganizationPhotos :execrows
INSERT INTO org_photos (organization_id, file_id, sort_order)
SELECT @organization_id::uuid, f.id, p.position::integer
FROM unnest(@file_ids::uuid[]) WITH ORDINALITY AS p (file_id, position)
JOIN files AS f ON f.id = p.file_id
WHERE f.organization_id = @organization_id::uuid AND f.purpose = 'STORE_PHOTO';

-- name: GetConsignorProfile :one
SELECT
    (delivery_radius_km IS NOT NULL)::boolean AS has_delivery_radius,
    COALESCE(delivery_radius_km, 0)::bigint AS delivery_radius_km,
    delivery_days, preferred_store_types, business_id_no, distribution_permit_no, halal_cert_no,
    can_provide_sample
FROM consignor_profiles
WHERE organization_id = @organization_id;

-- name: UpdateConsignorProfile :execrows
UPDATE consignor_profiles
SET
    delivery_radius_km = CASE WHEN @has_delivery_radius::boolean THEN @delivery_radius_km::bigint END,
    delivery_days = @delivery_days::text[],
    preferred_store_types = @preferred_store_types::text[],
    business_id_no = @business_id_no::text,
    distribution_permit_no = @distribution_permit_no::text,
    halal_cert_no = @halal_cert_no::text,
    can_provide_sample = @can_provide_sample::boolean,
    updated_at = @updated_at::timestamptz
WHERE organization_id = @organization_id::uuid;

-- name: GetPaymentInfo :one
SELECT
    o.payment_info, o.payment_methods, o.bank_name, o.bank_account_number, o.bank_account_holder,
    COALESCE(f.id::text, '')::text AS qr_file_id,
    COALESCE(f.storage_key, '')::text AS qr_storage_key
FROM organizations AS o
LEFT JOIN files AS f ON f.id = o.payment_qr_file_id
WHERE o.id = @id;

-- name: UpdatePaymentInfo :exec
UPDATE organizations
SET
    payment_info = @payment_info::text,
    payment_qr_file_id = NULLIF(@payment_qr_file_id::text, '')::uuid,
    payment_methods = @payment_methods::text[],
    bank_name = @bank_name::text,
    bank_account_number = @bank_account_number::text,
    bank_account_holder = @bank_account_holder::text,
    updated_at = @updated_at::timestamptz
WHERE id = @id::uuid;

-- name: GetListingFacts :one
SELECT
    o.type,
    o.is_listed,
    (o.name <> '')::boolean AS has_name,
    (o.city_code IS NOT NULL)::boolean AS has_city,
    (o.location IS NOT NULL)::boolean AS has_location,
    (o.description <> '')::boolean AS has_description,
    (o.logo_file_id IS NOT NULL)::boolean AS has_logo,
    (o.payment_info <> '' OR o.bank_account_number <> '' OR o.payment_qr_file_id IS NOT NULL)::boolean
        AS has_payment_info,
    COALESCE(sp.store_type IS NOT NULL, false)::boolean AS has_store_type,
    (SELECT count(*) FROM store_accepted_categories AS sac WHERE sac.organization_id = o.id)::bigint
        AS accepted_categories,
    (SELECT count(*) FROM org_photos AS p WHERE p.organization_id = o.id)::bigint AS photos,
    (SELECT count(*) FROM products AS pr WHERE pr.consignor_org_id = o.id AND NOT pr.is_archived)::bigint
        AS active_products
FROM organizations AS o
LEFT JOIN store_profiles AS sp ON sp.organization_id = o.id
WHERE o.id = @id;

-- name: SetOrganizationListed :exec
UPDATE organizations
SET is_listed = @is_listed::boolean, updated_at = @updated_at::timestamptz
WHERE id = @id::uuid;

-- name: SetOrganizationDemo :exec
UPDATE organizations SET is_demo = @is_demo::boolean WHERE id = @id::uuid;
