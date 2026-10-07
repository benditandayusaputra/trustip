-- +goose Up
CREATE TABLE organizations (
    id uuid PRIMARY KEY,
    type text NOT NULL,
    name text NOT NULL,
    slug text NOT NULL,
    description text NOT NULL DEFAULT '',
    whatsapp text NOT NULL,
    address text NOT NULL DEFAULT '',
    district text NOT NULL DEFAULT '',
    city_code text NOT NULL DEFAULT '',
    city_name text NOT NULL DEFAULT '',
    location geography(Point, 4326),
    location_public geography(Point, 4326),
    is_listed boolean NOT NULL DEFAULT false,
    status text NOT NULL DEFAULT 'ACTIVE',
    is_demo boolean NOT NULL DEFAULT false,
    last_active_at timestamptz,
    payment_info text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT organizations_slug_key UNIQUE (slug),
    CONSTRAINT organizations_type_known CHECK (type IN ('CONSIGNOR', 'STORE')),
    CONSTRAINT organizations_status_known CHECK (status IN ('ACTIVE', 'SUSPENDED')),
    CONSTRAINT organizations_slug_format CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
    CONSTRAINT organizations_whatsapp_format CHECK (whatsapp ~ '^62[0-9]{8,13}$'),
    CONSTRAINT organizations_location_pair CHECK ((location IS NULL) = (location_public IS NULL))
);

CREATE INDEX organizations_location_public_idx ON organizations USING gist (location_public);
CREATE INDEX organizations_name_trgm_idx ON organizations USING gin (name gin_trgm_ops);

-- +goose StatementBegin
CREATE FUNCTION reject_organization_type_change() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'organization type cannot change' USING ERRCODE = 'check_violation';
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER organizations_type_immutable
    BEFORE UPDATE OF type ON organizations
    FOR EACH ROW
    WHEN (OLD.type IS DISTINCT FROM NEW.type)
    EXECUTE FUNCTION reject_organization_type_change();

CREATE TABLE memberships (
    id uuid PRIMARY KEY,
    organization_id uuid NOT NULL REFERENCES organizations (id),
    user_id uuid NOT NULL REFERENCES users (id),
    role text NOT NULL,
    is_active boolean NOT NULL DEFAULT true,
    can_handle_payments boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT memberships_user_id_key UNIQUE (user_id),
    CONSTRAINT memberships_role_known CHECK (role IN ('OWNER', 'STAFF'))
);

CREATE INDEX memberships_organization_id_idx ON memberships (organization_id);

CREATE TABLE categories (
    id uuid PRIMARY KEY,
    name text NOT NULL,
    slug text NOT NULL,
    sort_order integer NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT categories_slug_key UNIQUE (slug),
    CONSTRAINT categories_slug_format CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$')
);

INSERT INTO categories (id, name, slug, sort_order) VALUES
    ('01a1161e-0a30-752d-ab9b-7397976f0dba', 'Roti dan kue basah', 'roti-dan-kue-basah', 1),
    ('01a1161e-0a49-71f7-a49a-3cf81042e3b0', 'Kue kering dan camilan', 'kue-kering-dan-camilan', 2),
    ('01a1161e-0a62-7a64-ab6f-024c2a8cc433', 'Keripik dan kerupuk', 'keripik-dan-kerupuk', 3),
    ('01a1161e-0a7a-7bbd-ac43-316028b3a3f8', 'Sambal dan bumbu', 'sambal-dan-bumbu', 4),
    ('01a1161e-0a92-70d6-9954-74cb98b0b1ac', 'Minuman', 'minuman', 5),
    ('01a1161e-0aac-7050-94ea-887eda363372', 'Makanan beku', 'makanan-beku', 6),
    ('01a1161e-0ac5-76c4-b0b4-50ee2cf17bea', 'Lauk siap saji', 'lauk-siap-saji', 7),
    ('01a1161e-0adf-746c-bf18-60b9149c3c40', 'Oleh-oleh khas daerah', 'oleh-oleh-khas-daerah', 8),
    ('01a1161e-0af8-78ed-b187-2cde3fd18697', 'Kerajinan dan suvenir', 'kerajinan-dan-suvenir', 9),
    ('01a1161e-0b10-79a0-b57b-9cf6de629684', 'Fesyen dan aksesori', 'fesyen-dan-aksesori', 10),
    ('01a1161e-0b29-7d88-861f-f07aa5ae7a84', 'Perawatan diri', 'perawatan-diri', 11),
    ('01a1161e-0b42-7391-b360-f3c2454cc118', 'Lainnya', 'lainnya', 12);

CREATE TABLE store_profiles (
    organization_id uuid PRIMARY KEY REFERENCES organizations (id),
    store_type text,
    opening_hours text NOT NULL DEFAULT '',
    accepts_fixed_deposit boolean NOT NULL DEFAULT false,
    accepts_commission boolean NOT NULL DEFAULT false,
    commission_min_bps integer,
    commission_max_bps integer,
    requires_distribution_permit boolean NOT NULL DEFAULT false,
    requires_halal boolean NOT NULL DEFAULT false,
    requires_sample boolean NOT NULL DEFAULT false,
    other_requirements text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT store_profiles_store_type_known CHECK (store_type IN (
        'GROCERY_STALL', 'LOCAL_MINIMARKET', 'SOUVENIR_SHOP', 'CANTEEN', 'CAFE',
        'BAKERY', 'COOPERATIVE', 'MARKET_KIOSK', 'OTHER'
    )),
    CONSTRAINT store_profiles_commission_range CHECK (
        commission_min_bps BETWEEN 0 AND 5000
        AND commission_max_bps BETWEEN 0 AND 5000
        AND commission_min_bps <= commission_max_bps
    )
);

CREATE INDEX store_profiles_store_type_idx ON store_profiles (store_type);

CREATE TABLE consignor_profiles (
    organization_id uuid PRIMARY KEY REFERENCES organizations (id),
    delivery_radius_km integer,
    delivery_days text[] NOT NULL DEFAULT '{}',
    preferred_store_types text[] NOT NULL DEFAULT '{}',
    business_id_no text NOT NULL DEFAULT '',
    distribution_permit_no text NOT NULL DEFAULT '',
    halal_cert_no text NOT NULL DEFAULT '',
    can_provide_sample boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT consignor_profiles_delivery_radius_positive CHECK (delivery_radius_km > 0),
    CONSTRAINT consignor_profiles_delivery_days_known CHECK (
        delivery_days <@ ARRAY['MON', 'TUE', 'WED', 'THU', 'FRI', 'SAT', 'SUN']
    ),
    CONSTRAINT consignor_profiles_preferred_store_types_known CHECK (preferred_store_types <@ ARRAY[
        'GROCERY_STALL', 'LOCAL_MINIMARKET', 'SOUVENIR_SHOP', 'CANTEEN', 'CAFE',
        'BAKERY', 'COOPERATIVE', 'MARKET_KIOSK', 'OTHER'
    ])
);

CREATE TABLE store_accepted_categories (
    organization_id uuid NOT NULL REFERENCES organizations (id),
    category_id uuid NOT NULL REFERENCES categories (id),
    PRIMARY KEY (organization_id, category_id)
);

CREATE INDEX store_accepted_categories_category_id_idx ON store_accepted_categories (category_id);

-- +goose Down
DROP TABLE store_accepted_categories;
DROP TABLE consignor_profiles;
DROP TABLE store_profiles;
DROP TABLE categories;
DROP TABLE memberships;
DROP TABLE organizations;
DROP FUNCTION reject_organization_type_change();
