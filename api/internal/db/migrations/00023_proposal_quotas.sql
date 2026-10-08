-- +goose Up
CREATE TABLE proposal_quotas (
    organization_id uuid NOT NULL REFERENCES organizations (id),
    day date NOT NULL,
    sent integer NOT NULL,
    PRIMARY KEY (organization_id, day),
    CONSTRAINT proposal_quotas_daily_limit CHECK (sent BETWEEN 1 AND 10)
);

CREATE INDEX proposals_declined_from_idx ON proposals (from_org_id, decided_at DESC) WHERE status = 'DECLINED';

-- +goose Down
DROP INDEX proposals_declined_from_idx;
DROP TABLE proposal_quotas;
