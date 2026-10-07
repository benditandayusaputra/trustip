-- +goose Up
CREATE TABLE org_photos (
    organization_id uuid NOT NULL REFERENCES organizations (id),
    file_id uuid NOT NULL REFERENCES files (id),
    sort_order integer NOT NULL,
    PRIMARY KEY (organization_id, file_id),
    CONSTRAINT org_photos_sort_order_key UNIQUE (organization_id, sort_order),
    CONSTRAINT org_photos_at_most_five CHECK (sort_order BETWEEN 1 AND 5)
);

-- +goose Down
DROP TABLE org_photos;
