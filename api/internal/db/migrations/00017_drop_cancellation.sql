-- +goose Up
ALTER TABLE drops ADD COLUMN cancelled_by uuid REFERENCES users (id);
ALTER TABLE drops ADD COLUMN cancelled_at timestamptz;

-- +goose Down
ALTER TABLE drops DROP COLUMN cancelled_at;
ALTER TABLE drops DROP COLUMN cancelled_by;
