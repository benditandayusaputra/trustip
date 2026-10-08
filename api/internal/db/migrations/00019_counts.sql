-- +goose Up
CREATE TABLE counts (
    id uuid PRIMARY KEY,
    number text NOT NULL,
    partnership_id uuid NOT NULL REFERENCES partnerships (id),
    store_org_id uuid NOT NULL REFERENCES organizations (id),
    occasion text NOT NULL,
    counted_at timestamptz NOT NULL,
    status text NOT NULL DEFAULT 'RECORDED',
    note text NOT NULL DEFAULT '',
    sold_qty integer NOT NULL,
    returned_qty integer NOT NULL,
    gross_amount bigint NOT NULL,
    store_share_amount bigint NOT NULL,
    consignor_share_amount bigint NOT NULL,
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
    CONSTRAINT counts_number_key UNIQUE (store_org_id, number),
    CONSTRAINT counts_idempotency_key UNIQUE (store_org_id, idempotency_key),
    CONSTRAINT counts_occasion_known CHECK (occasion IN ('STORE_CLOSING', 'PICKUP', 'BEFORE_DROP', 'OTHER')),
    CONSTRAINT counts_status_known CHECK (status IN ('RECORDED', 'CONFIRMED', 'DISPUTED')),
    CONSTRAINT counts_quantities_not_negative CHECK (sold_qty >= 0 AND returned_qty >= 0),
    CONSTRAINT counts_split_adds_up CHECK (
        gross_amount >= 0 AND store_share_amount + consignor_share_amount = gross_amount
    ),
    CONSTRAINT counts_fingerprint_length CHECK (octet_length(request_fingerprint) = 32)
);

CREATE INDEX counts_partnership_counted_idx ON counts (partnership_id, counted_at DESC, id DESC);

CREATE TABLE count_lines (
    id uuid PRIMARY KEY,
    count_id uuid NOT NULL REFERENCES counts (id),
    line_no integer NOT NULL,
    product_id uuid NOT NULL REFERENCES products (id),
    recorded_qty integer NOT NULL,
    counted_qty integer NOT NULL,
    sold_qty integer NOT NULL,
    returned_qty integer NOT NULL,
    kept_qty integer NOT NULL,
    return_reason text NOT NULL DEFAULT '',
    CONSTRAINT count_lines_line_key UNIQUE (count_id, line_no),
    CONSTRAINT count_lines_product_key UNIQUE (count_id, product_id),
    CONSTRAINT count_lines_counted_range CHECK (counted_qty BETWEEN 0 AND recorded_qty),
    CONSTRAINT count_lines_returned_range CHECK (returned_qty BETWEEN 0 AND counted_qty),
    CONSTRAINT count_lines_adds_up CHECK (
        sold_qty = recorded_qty - counted_qty AND kept_qty = counted_qty - returned_qty
    ),
    CONSTRAINT count_lines_reason_known CHECK (
        (returned_qty = 0 AND return_reason = '')
        OR (returned_qty > 0 AND return_reason IN ('UNSOLD', 'EXPIRED', 'NEAR_EXPIRY', 'DAMAGED', 'RECALL', 'OTHER'))
    )
);

CREATE TABLE sale_allocations (
    id uuid PRIMARY KEY,
    count_line_id uuid NOT NULL REFERENCES count_lines (id),
    drop_item_id uuid NOT NULL REFERENCES drop_items (id),
    partnership_id uuid NOT NULL REFERENCES partnerships (id),
    product_id uuid NOT NULL REFERENCES products (id),
    qty integer NOT NULL,
    gross_amount bigint NOT NULL,
    store_share_amount bigint NOT NULL,
    consignor_share_amount bigint NOT NULL,
    occurred_at timestamptz NOT NULL,
    settlement_id uuid,
    CONSTRAINT sale_allocations_qty_positive CHECK (qty > 0),
    CONSTRAINT sale_allocations_split_adds_up CHECK (
        gross_amount >= 0 AND store_share_amount + consignor_share_amount = gross_amount
    )
);

CREATE INDEX sale_allocations_partnership_occurred_idx ON sale_allocations (partnership_id, occurred_at);
CREATE INDEX sale_allocations_settlement_idx ON sale_allocations (settlement_id);

CREATE TABLE return_allocations (
    id uuid PRIMARY KEY,
    count_line_id uuid NOT NULL REFERENCES count_lines (id),
    drop_item_id uuid NOT NULL REFERENCES drop_items (id),
    qty integer NOT NULL,
    CONSTRAINT return_allocations_qty_positive CHECK (qty > 0)
);

CREATE INDEX return_allocations_drop_item_idx ON return_allocations (drop_item_id);

-- +goose Down
DROP TABLE return_allocations;
DROP TABLE sale_allocations;
DROP TABLE count_lines;
DROP TABLE counts;
