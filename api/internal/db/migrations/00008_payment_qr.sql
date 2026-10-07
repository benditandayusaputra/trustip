-- +goose Up
ALTER TABLE organizations ADD COLUMN payment_qr_file_id uuid REFERENCES files (id);

-- +goose Down
ALTER TABLE organizations DROP COLUMN payment_qr_file_id;
