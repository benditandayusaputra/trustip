-- +goose Up
ALTER TABLE proposals ADD COLUMN decision_reason text NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE proposals DROP COLUMN decision_reason;
