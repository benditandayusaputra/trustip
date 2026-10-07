-- +goose Up
CREATE TABLE seals (
    id uuid PRIMARY KEY,
    partnership_id uuid NOT NULL REFERENCES partnerships (id),
    seq integer NOT NULL,
    format_version integer NOT NULL,
    event_type text NOT NULL,
    entity_type text NOT NULL,
    entity_id uuid NOT NULL,
    entity_number text NOT NULL,
    content_hash bytea NOT NULL,
    actor_user_id uuid NOT NULL REFERENCES users (id),
    actor_org_id uuid NOT NULL REFERENCES organizations (id),
    occurred_at timestamptz NOT NULL,
    prev_hash bytea NOT NULL,
    entry_hash bytea NOT NULL,
    signature bytea NOT NULL,
    key_id text NOT NULL,
    code text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT seals_seq_key UNIQUE (partnership_id, seq),
    CONSTRAINT seals_code_key UNIQUE (code),
    CONSTRAINT seals_seq_positive CHECK (seq > 0),
    CONSTRAINT seals_event_type_format CHECK (event_type ~ '^[A-Z_]+$'),
    CONSTRAINT seals_entity_type_format CHECK (entity_type ~ '^[A-Z_]+$'),
    CONSTRAINT seals_content_hash_length CHECK (octet_length(content_hash) = 32),
    CONSTRAINT seals_entry_hash_length CHECK (octet_length(entry_hash) = 32),
    CONSTRAINT seals_prev_hash_length CHECK (octet_length(prev_hash) IN (0, 32)),
    CONSTRAINT seals_signature_length CHECK (octet_length(signature) = 64),
    CONSTRAINT seals_code_format CHECK (code ~ '^[A-Z2-7]{4}(-[A-Z2-7]{4}){3}$')
);

CREATE INDEX seals_entity_idx ON seals (entity_type, entity_id, seq DESC);

-- +goose StatementBegin
CREATE FUNCTION reject_seal_change() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    IF TG_OP = 'DELETE' AND EXISTS (
        SELECT 1
        FROM partnerships AS p
        JOIN organizations AS c ON c.id = p.consignor_org_id
        JOIN organizations AS s ON s.id = p.store_org_id
        WHERE p.id = OLD.partnership_id AND c.is_demo AND s.is_demo
    ) THEN
        RETURN OLD;
    END IF;
    RAISE EXCEPTION 'seals are append only' USING ERRCODE = 'restrict_violation';
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER seals_append_only
    BEFORE UPDATE OR DELETE ON seals
    FOR EACH ROW
    EXECUTE FUNCTION reject_seal_change();

-- +goose StatementBegin
CREATE FUNCTION reject_seal_truncate() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'seals are append only' USING ERRCODE = 'restrict_violation';
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER seals_no_truncate
    BEFORE TRUNCATE ON seals
    FOR EACH STATEMENT
    EXECUTE FUNCTION reject_seal_truncate();

-- +goose Down
DROP TABLE seals;
DROP FUNCTION reject_seal_truncate();
DROP FUNCTION reject_seal_change();
