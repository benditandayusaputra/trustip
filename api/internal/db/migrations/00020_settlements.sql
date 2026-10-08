-- +goose Up
CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE settlements (
    id uuid PRIMARY KEY,
    number text NOT NULL,
    partnership_id uuid NOT NULL REFERENCES partnerships (id),
    consignor_org_id uuid NOT NULL REFERENCES organizations (id),
    store_org_id uuid NOT NULL REFERENCES organizations (id),
    schedule text NOT NULL,
    period_start timestamptz NOT NULL,
    period_end timestamptz NOT NULL,
    status text NOT NULL DEFAULT 'ISSUED',
    sold_qty integer NOT NULL,
    gross_amount bigint NOT NULL,
    store_share_amount bigint NOT NULL,
    consignor_share_amount bigint NOT NULL,
    due_date date NOT NULL,
    issued_at timestamptz NOT NULL,
    paid_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT settlements_number_key UNIQUE (consignor_org_id, number),
    CONSTRAINT settlements_schedule_known CHECK (schedule IN ('ON_COUNT', 'WEEKLY', 'MONTHLY')),
    CONSTRAINT settlements_status_known CHECK (status IN ('ISSUED', 'PAID')),
    CONSTRAINT settlements_paid_has_time CHECK ((status = 'PAID') = (paid_at IS NOT NULL)),
    CONSTRAINT settlements_period_ordered CHECK (period_start < period_end),
    CONSTRAINT settlements_sold_positive CHECK (sold_qty > 0),
    CONSTRAINT settlements_split_adds_up CHECK (
        gross_amount >= 0 AND store_share_amount + consignor_share_amount = gross_amount
    ),
    CONSTRAINT settlements_periods_disjoint EXCLUDE USING gist (
        partnership_id WITH =, tstzrange(period_start, period_end) WITH &&
    )
);

CREATE INDEX settlements_partnership_status_due_idx ON settlements (partnership_id, status, due_date);
CREATE INDEX settlements_consignor_issued_idx ON settlements (consignor_org_id, issued_at DESC, id DESC);
CREATE INDEX settlements_store_issued_idx ON settlements (store_org_id, issued_at DESC, id DESC);
CREATE INDEX sale_allocations_unsettled_idx ON sale_allocations (partnership_id, occurred_at)
    WHERE settlement_id IS NULL;

ALTER TABLE sale_allocations
    ADD CONSTRAINT sale_allocations_settlement_fk FOREIGN KEY (settlement_id) REFERENCES settlements (id);

ALTER TABLE seals ALTER COLUMN actor_user_id DROP NOT NULL;
ALTER TABLE seals ALTER COLUMN actor_org_id DROP NOT NULL;
ALTER TABLE audit_logs ALTER COLUMN actor_id DROP NOT NULL;

-- +goose Down
ALTER TABLE audit_logs ALTER COLUMN actor_id SET NOT NULL;
ALTER TABLE seals ALTER COLUMN actor_org_id SET NOT NULL;
ALTER TABLE seals ALTER COLUMN actor_user_id SET NOT NULL;
ALTER TABLE sale_allocations DROP CONSTRAINT sale_allocations_settlement_fk;
DROP INDEX sale_allocations_unsettled_idx;
DROP TABLE settlements;
