-- +goose Up
CREATE TABLE notifications (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users (id),
    type text NOT NULL,
    title text NOT NULL,
    body text NOT NULL,
    link text NOT NULL,
    read_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT notifications_type_format CHECK (type ~ '^[A-Z_]+$'),
    CONSTRAINT notifications_link_in_app CHECK (link = '/app' OR link LIKE '/app/%')
);

CREATE INDEX notifications_user_id_idx ON notifications (user_id, created_at DESC, id DESC);
CREATE INDEX notifications_unread_idx ON notifications (user_id) WHERE read_at IS NULL;

-- +goose Down
DROP TABLE notifications;
