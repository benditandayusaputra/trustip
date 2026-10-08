-- +goose Up
CREATE TABLE listing_reports (
    id uuid PRIMARY KEY,
    reporter_user_id uuid NOT NULL REFERENCES users (id),
    reporter_org_id uuid NOT NULL REFERENCES organizations (id),
    target_type text NOT NULL,
    target_id uuid NOT NULL,
    reason text NOT NULL,
    note text NOT NULL DEFAULT '',
    status text NOT NULL DEFAULT 'OPEN',
    resolved_by uuid REFERENCES users (id),
    resolved_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT listing_reports_target_known CHECK (target_type IN ('ORGANIZATION', 'PRODUCT', 'NEED')),
    CONSTRAINT listing_reports_reason_known CHECK (
        reason IN ('MISLEADING', 'PROHIBITED_GOODS', 'OFFENSIVE', 'SPAM', 'OTHER')
    ),
    CONSTRAINT listing_reports_note_length CHECK (char_length(note) <= 500),
    CONSTRAINT listing_reports_status_known CHECK (status IN ('OPEN', 'RESOLVED', 'DISMISSED')),
    CONSTRAINT listing_reports_resolution_complete CHECK ((status = 'OPEN') = (resolved_at IS NULL))
);

CREATE INDEX listing_reports_status_idx ON listing_reports (status, created_at);
CREATE UNIQUE INDEX listing_reports_open_key ON listing_reports (reporter_user_id, target_type, target_id)
    WHERE status = 'OPEN';

-- +goose Down
DROP TABLE listing_reports;
