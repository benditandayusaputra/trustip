-- +goose Up
CREATE TABLE org_stats (
    organization_id uuid PRIMARY KEY REFERENCES organizations (id),
    active_partner_count integer NOT NULL DEFAULT 0,
    on_time_payment_bps integer NOT NULL DEFAULT 0,
    settlements_counted integer NOT NULL DEFAULT 0,
    product_count integer NOT NULL DEFAULT 0,
    computed_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT org_stats_counts_not_negative CHECK (
        active_partner_count >= 0 AND settlements_counted >= 0 AND product_count >= 0
    ),
    CONSTRAINT org_stats_on_time_range CHECK (on_time_payment_bps BETWEEN 0 AND 10000)
);

-- +goose StatementBegin
CREATE FUNCTION mark_organization_active() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    UPDATE organizations AS o
    SET last_active_at = NEW.created_at
    FROM memberships AS m
    WHERE m.user_id = NEW.user_id AND m.is_active AND o.id = m.organization_id
        AND (o.last_active_at IS NULL OR o.last_active_at < NEW.created_at);
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER sessions_mark_organization_active
    AFTER INSERT ON sessions
    FOR EACH ROW
    EXECUTE FUNCTION mark_organization_active();

CREATE INDEX organizations_listed_type_idx ON organizations (type, created_at DESC, id DESC)
    WHERE status = 'ACTIVE' AND is_listed;

-- +goose Down
DROP INDEX organizations_listed_type_idx;
DROP TRIGGER sessions_mark_organization_active ON sessions;
DROP FUNCTION mark_organization_active();
DROP TABLE org_stats;
