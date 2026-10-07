-- +goose Up
ALTER TABLE organizations ADD COLUMN logo_file_id uuid REFERENCES files (id);

-- +goose Down
ALTER TABLE organizations DROP COLUMN logo_file_id;
