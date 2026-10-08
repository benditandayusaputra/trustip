-- +goose Up
CREATE TABLE needs (
    id uuid PRIMARY KEY,
    store_org_id uuid NOT NULL REFERENCES organizations (id),
    title text NOT NULL,
    category_id uuid NOT NULL REFERENCES categories (id),
    description text NOT NULL DEFAULT '',
    est_qty_per_week integer,
    max_retail_price bigint,
    preferred_model text,
    offered_commission_bps integer,
    status text NOT NULL DEFAULT 'OPEN',
    expires_at timestamptz NOT NULL,
    reminded_at timestamptz,
    created_by uuid NOT NULL REFERENCES users (id),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT needs_status_known CHECK (status IN ('OPEN', 'FULFILLED', 'CLOSED', 'EXPIRED')),
    CONSTRAINT needs_title_length CHECK (char_length(title) BETWEEN 5 AND 80),
    CONSTRAINT needs_description_length CHECK (char_length(description) <= 500),
    CONSTRAINT needs_quantity_positive CHECK (est_qty_per_week > 0),
    CONSTRAINT needs_price_positive CHECK (max_retail_price > 0),
    CONSTRAINT needs_model_known CHECK (preferred_model IN ('FIXED_DEPOSIT', 'COMMISSION')),
    CONSTRAINT needs_commission_range CHECK (offered_commission_bps BETWEEN 0 AND 5000)
);

CREATE INDEX needs_status_category_idx ON needs (status, category_id);
CREATE INDEX needs_store_org_id_idx ON needs (store_org_id, created_at DESC, id DESC);

-- +goose StatementBegin
CREATE FUNCTION limit_open_needs() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    PERFORM 1 FROM organizations WHERE id = NEW.store_org_id FOR UPDATE;
    IF (SELECT count(*) FROM needs
        WHERE store_org_id = NEW.store_org_id AND status = 'OPEN' AND id <> NEW.id) >= 5 THEN
        RAISE EXCEPTION 'a store may have at most 5 open needs'
            USING ERRCODE = 'check_violation', CONSTRAINT = 'needs_open_limit';
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER needs_open_limit
    BEFORE INSERT OR UPDATE OF status ON needs
    FOR EACH ROW
    WHEN (NEW.status = 'OPEN')
    EXECUTE FUNCTION limit_open_needs();

ALTER TABLE proposals ADD COLUMN need_id uuid REFERENCES needs (id);
CREATE INDEX proposals_need_id_idx ON proposals (need_id) WHERE need_id IS NOT NULL;

-- +goose Down
DROP INDEX proposals_need_id_idx;
ALTER TABLE proposals DROP COLUMN need_id;
DROP TRIGGER needs_open_limit ON needs;
DROP FUNCTION limit_open_needs();
DROP TABLE needs;
