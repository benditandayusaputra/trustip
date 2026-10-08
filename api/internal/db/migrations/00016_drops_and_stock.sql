-- +goose Up
CREATE TABLE drops (
    id uuid PRIMARY KEY,
    number text NOT NULL,
    partnership_id uuid NOT NULL REFERENCES partnerships (id),
    agreement_id uuid NOT NULL REFERENCES agreements (id),
    store_org_id uuid NOT NULL REFERENCES organizations (id),
    initiated_by text NOT NULL DEFAULT 'STORE',
    status text NOT NULL DEFAULT 'RECORDED',
    occurred_at timestamptz NOT NULL,
    note text NOT NULL DEFAULT '',
    recorded_by uuid NOT NULL REFERENCES users (id),
    confirmed_by uuid REFERENCES users (id),
    confirmed_at timestamptz,
    auto_confirmed boolean NOT NULL DEFAULT false,
    idempotency_key uuid NOT NULL,
    request_fingerprint bytea NOT NULL,
    client_recorded_at timestamptz NOT NULL,
    received_at timestamptz NOT NULL,
    recorded_offline boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT drops_number_key UNIQUE (store_org_id, number),
    CONSTRAINT drops_idempotency_key UNIQUE (store_org_id, idempotency_key),
    CONSTRAINT drops_initiated_by_known CHECK (initiated_by IN ('STORE', 'CONSIGNOR')),
    CONSTRAINT drops_status_known CHECK (status IN ('SENT', 'RECORDED', 'CONFIRMED', 'DISPUTED', 'CANCELLED')),
    CONSTRAINT drops_fingerprint_length CHECK (octet_length(request_fingerprint) = 32)
);

CREATE INDEX drops_partnership_occurred_idx ON drops (partnership_id, occurred_at DESC, id DESC);

CREATE TABLE drop_items (
    id uuid PRIMARY KEY,
    drop_id uuid NOT NULL REFERENCES drops (id),
    line_no integer NOT NULL,
    product_id uuid NOT NULL REFERENCES products (id),
    batch_code text NOT NULL,
    expiry_date date,
    qty_declared integer NOT NULL,
    qty_received integer NOT NULL,
    pricing_model text NOT NULL,
    retail_price bigint NOT NULL,
    commission_bps integer NOT NULL DEFAULT 0,
    deposit_price bigint NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT drop_items_line_key UNIQUE (drop_id, line_no),
    CONSTRAINT drop_items_qty_range CHECK (qty_received BETWEEN 1 AND 10000 AND qty_declared BETWEEN 1 AND 10000),
    CONSTRAINT drop_items_model_known CHECK (pricing_model IN ('FIXED_DEPOSIT', 'COMMISSION')),
    CONSTRAINT drop_items_prices_range CHECK (
        retail_price BETWEEN 0 AND 100000000 AND deposit_price BETWEEN 0 AND retail_price
    ),
    CONSTRAINT drop_items_commission_range CHECK (commission_bps BETWEEN 0 AND 5000)
);

CREATE INDEX drop_items_product_expiry_idx ON drop_items (product_id, expiry_date);

CREATE TABLE stock_movements (
    id uuid PRIMARY KEY,
    partnership_id uuid NOT NULL REFERENCES partnerships (id),
    product_id uuid NOT NULL REFERENCES products (id),
    drop_item_id uuid NOT NULL REFERENCES drop_items (id),
    type text NOT NULL,
    qty integer NOT NULL,
    occurred_at timestamptz NOT NULL,
    source_type text NOT NULL,
    source_id uuid NOT NULL,
    source_number text NOT NULL,
    created_by uuid NOT NULL REFERENCES users (id),
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT stock_movements_type_known CHECK (type IN ('RECEIVE', 'SALE', 'RETURN', 'ADJUSTMENT')),
    CONSTRAINT stock_movements_sign CHECK (
        (type = 'RECEIVE' AND qty > 0) OR (type IN ('SALE', 'RETURN') AND qty < 0)
        OR (type = 'ADJUSTMENT' AND qty <> 0)
    ),
    CONSTRAINT stock_movements_source_known CHECK (source_type IN ('DROP', 'COUNT'))
);

CREATE INDEX stock_movements_partnership_product_idx ON stock_movements (partnership_id, product_id, occurred_at);
CREATE INDEX stock_movements_drop_item_idx ON stock_movements (drop_item_id);

-- +goose Down
DROP TABLE stock_movements;
DROP TABLE drop_items;
DROP TABLE drops;
