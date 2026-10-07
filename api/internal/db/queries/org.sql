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

-- name: FindOrganizationLocation :one
SELECT
    ST_Y(location::geometry)::float8 AS latitude,
    ST_X(location::geometry)::float8 AS longitude,
    ST_Y(location_public::geometry)::float8 AS public_latitude,
    ST_X(location_public::geometry)::float8 AS public_longitude
FROM organizations
WHERE id = @id AND location IS NOT NULL;
