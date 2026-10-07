-- +goose Up
ALTER TABLE products ADD COLUMN is_seeking_stores boolean NOT NULL DEFAULT false;

CREATE INDEX products_seeking_category_id_idx ON products (category_id)
    WHERE is_seeking_stores AND NOT is_archived;

-- +goose Down
DROP INDEX products_seeking_category_id_idx;
ALTER TABLE products DROP COLUMN is_seeking_stores;
