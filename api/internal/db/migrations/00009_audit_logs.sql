-- +goose Up
CREATE TABLE audit_logs (
    id uuid PRIMARY KEY,
    organization_id uuid REFERENCES organizations (id),
    actor_id uuid NOT NULL REFERENCES users (id),
    action text NOT NULL,
    entity_type text NOT NULL,
    entity_id uuid NOT NULL,
    metadata jsonb NOT NULL DEFAULT '{}',
    ip text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT audit_logs_action_format CHECK (action ~ '^[A-Z_]+$'),
    CONSTRAINT audit_logs_entity_type_format CHECK (entity_type ~ '^[A-Z_]+$')
);

CREATE INDEX audit_logs_organization_id_idx ON audit_logs (organization_id, created_at);

-- +goose Down
DROP TABLE audit_logs;
