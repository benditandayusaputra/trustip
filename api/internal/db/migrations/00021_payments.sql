-- +goose Up
ALTER TABLE files DROP CONSTRAINT files_purpose_known;
ALTER TABLE files ADD CONSTRAINT files_purpose_known
    CHECK (purpose IN ('LOGO', 'STORE_PHOTO', 'PAYMENT_QR', 'PRODUCT_PHOTO', 'PAYMENT_PROOF'));

CREATE TABLE payments (
    id uuid PRIMARY KEY,
    settlement_id uuid NOT NULL REFERENCES settlements (id),
    partnership_id uuid NOT NULL REFERENCES partnerships (id),
    amount bigint NOT NULL,
    method text NOT NULL,
    paid_on date NOT NULL,
    proof_file_id uuid REFERENCES files (id),
    proof_sha256 bytea,
    status text NOT NULL DEFAULT 'PENDING_CONFIRMATION',
    rejection_reason text NOT NULL DEFAULT '',
    recorded_by uuid NOT NULL REFERENCES users (id),
    recorded_by_org_id uuid NOT NULL REFERENCES organizations (id),
    reviewed_by uuid REFERENCES users (id),
    reviewed_at timestamptz,
    idempotency_key uuid NOT NULL,
    request_fingerprint bytea NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT payments_idempotency_key UNIQUE (recorded_by_org_id, idempotency_key),
    CONSTRAINT payments_amount_positive CHECK (amount > 0),
    CONSTRAINT payments_method_known CHECK (method IN ('TRANSFER', 'QRIS', 'CASH')),
    CONSTRAINT payments_status_known CHECK (status IN ('PENDING_CONFIRMATION', 'CONFIRMED', 'REJECTED')),
    CONSTRAINT payments_proof_for_transfers CHECK (method = 'CASH' OR proof_file_id IS NOT NULL),
    CONSTRAINT payments_proof_hashed CHECK ((proof_file_id IS NULL) = (proof_sha256 IS NULL)),
    CONSTRAINT payments_rejection_has_reason CHECK ((status = 'REJECTED') = (rejection_reason <> '')),
    CONSTRAINT payments_fingerprint_length CHECK (octet_length(request_fingerprint) = 32)
);

CREATE INDEX payments_settlement_status_idx ON payments (settlement_id, status);
CREATE UNIQUE INDEX payments_one_open_key ON payments (settlement_id)
    WHERE status IN ('PENDING_CONFIRMATION', 'CONFIRMED');
CREATE INDEX payments_proof_file_idx ON payments (proof_file_id) WHERE proof_file_id IS NOT NULL;

-- +goose Down
DROP TABLE payments;
ALTER TABLE files DROP CONSTRAINT files_purpose_known;
ALTER TABLE files ADD CONSTRAINT files_purpose_known
    CHECK (purpose IN ('LOGO', 'STORE_PHOTO', 'PAYMENT_QR', 'PRODUCT_PHOTO'));
