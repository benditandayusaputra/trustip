-- +goose Up
ALTER TABLE files DROP CONSTRAINT files_purpose_known;
ALTER TABLE files ADD CONSTRAINT files_purpose_known
    CHECK (purpose IN ('LOGO', 'STORE_PHOTO', 'PAYMENT_QR', 'PRODUCT_PHOTO'));

CREATE TABLE products (
    id uuid PRIMARY KEY,
    consignor_org_id uuid NOT NULL REFERENCES organizations (id),
    sku text NOT NULL,
    name text NOT NULL,
    category_id uuid NOT NULL REFERENCES categories (id),
    unit text NOT NULL,
    description text NOT NULL DEFAULT '',
    composition text NOT NULL DEFAULT '',
    net_content text NOT NULL DEFAULT '',
    photo_file_id uuid REFERENCES files (id),
    suggested_retail_price bigint NOT NULL,
    suggested_deposit_price bigint NOT NULL,
    shelf_life_days integer,
    is_archived boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT products_sku_key UNIQUE (consignor_org_id, sku),
    CONSTRAINT products_sku_format CHECK (sku ~ '^[A-Z0-9]+(-[A-Z0-9]+)*$'),
    CONSTRAINT products_suggested_prices_range CHECK (
        suggested_retail_price BETWEEN 0 AND 100000000
        AND suggested_deposit_price BETWEEN 0 AND suggested_retail_price
    ),
    CONSTRAINT products_shelf_life_range CHECK (shelf_life_days BETWEEN 1 AND 730)
);

CREATE INDEX products_consignor_org_id_idx ON products (consignor_org_id, created_at DESC, id DESC);
CREATE INDEX products_name_trgm_idx ON products USING gin (name gin_trgm_ops);

-- +goose Down
DROP TABLE products;
ALTER TABLE files DROP CONSTRAINT files_purpose_known;
ALTER TABLE files ADD CONSTRAINT files_purpose_known CHECK (purpose IN ('LOGO', 'STORE_PHOTO', 'PAYMENT_QR'));
