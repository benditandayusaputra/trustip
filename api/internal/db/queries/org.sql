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
