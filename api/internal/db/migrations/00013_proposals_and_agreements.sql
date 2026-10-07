-- +goose Up
CREATE TABLE document_sequences (
    organization_id uuid NOT NULL REFERENCES organizations (id),
    prefix text NOT NULL,
    period text NOT NULL,
    last_value integer NOT NULL,
    PRIMARY KEY (organization_id, prefix, period),
    CONSTRAINT document_sequences_last_value_positive CHECK (last_value > 0)
);

CREATE TABLE partnerships (
    id uuid PRIMARY KEY,
    consignor_org_id uuid NOT NULL REFERENCES organizations (id),
    store_org_id uuid NOT NULL REFERENCES organizations (id),
    status text NOT NULL DEFAULT 'ACTIVE',
    show_on_public boolean NOT NULL DEFAULT false,
    started_at timestamptz NOT NULL,
    ended_at timestamptz,
    ended_by_org_id uuid REFERENCES organizations (id),
    end_reason text NOT NULL DEFAULT '',
    seal_seq integer NOT NULL DEFAULT 0,
    seal_head_hash bytea,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT partnerships_status_known CHECK (status IN ('ACTIVE', 'ENDED')),
    CONSTRAINT partnerships_two_parties CHECK (consignor_org_id <> store_org_id),
    CONSTRAINT partnerships_seal_seq_not_negative CHECK (seal_seq >= 0)
);

CREATE UNIQUE INDEX partnerships_active_pair_key ON partnerships (consignor_org_id, store_org_id)
    WHERE status = 'ACTIVE';
CREATE INDEX partnerships_store_org_id_idx ON partnerships (store_org_id);

CREATE TABLE proposals (
    id uuid PRIMARY KEY,
    number text NOT NULL,
    kind text NOT NULL,
    consignor_org_id uuid REFERENCES organizations (id),
    store_org_id uuid REFERENCES organizations (id),
    from_org_id uuid NOT NULL REFERENCES organizations (id),
    awaiting_org_id uuid REFERENCES organizations (id),
    partnership_id uuid REFERENCES partnerships (id),
    invite_code text,
    status text NOT NULL DEFAULT 'OPEN',
    current_revision_no integer NOT NULL DEFAULT 1,
    expires_at timestamptz NOT NULL,
    decided_by uuid REFERENCES users (id),
    decided_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT proposals_number_key UNIQUE (from_org_id, number),
    CONSTRAINT proposals_invite_code_key UNIQUE (invite_code),
    CONSTRAINT proposals_kind_known CHECK (kind IN ('NEW', 'AMENDMENT')),
    CONSTRAINT proposals_status_known CHECK (status IN ('OPEN', 'ACCEPTED', 'DECLINED', 'WITHDRAWN', 'EXPIRED')),
    CONSTRAINT proposals_invite_code_format CHECK (invite_code ~ '^[A-HJ-NP-Z2-9]{10}$'),
    CONSTRAINT proposals_sender_is_a_party CHECK (
        (from_org_id = consignor_org_id) IS TRUE OR (from_org_id = store_org_id) IS TRUE
    ),
    CONSTRAINT proposals_two_parties CHECK (consignor_org_id <> store_org_id),
    CONSTRAINT proposals_revision_no_positive CHECK (current_revision_no > 0)
);

CREATE UNIQUE INDEX proposals_open_pair_key ON proposals (consignor_org_id, store_org_id, kind)
    WHERE status = 'OPEN';
CREATE INDEX proposals_consignor_org_id_idx ON proposals (consignor_org_id, created_at DESC, id DESC);
CREATE INDEX proposals_store_org_id_idx ON proposals (store_org_id, created_at DESC, id DESC);

CREATE TABLE proposal_revisions (
    id uuid PRIMARY KEY,
    proposal_id uuid NOT NULL REFERENCES proposals (id),
    revision_no integer NOT NULL,
    author_org_id uuid NOT NULL REFERENCES organizations (id),
    author_user_id uuid NOT NULL REFERENCES users (id),
    settlement_schedule text NOT NULL,
    payment_terms_days integer NOT NULL,
    count_rhythm text NOT NULL,
    notes text NOT NULL DEFAULT '',
    message text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT proposal_revisions_number_key UNIQUE (proposal_id, revision_no),
    CONSTRAINT proposal_revisions_schedule_known CHECK (settlement_schedule IN ('ON_COUNT', 'WEEKLY', 'MONTHLY')),
    CONSTRAINT proposal_revisions_payment_terms_range CHECK (payment_terms_days BETWEEN 0 AND 30),
    CONSTRAINT proposal_revisions_rhythm_known CHECK (count_rhythm IN ('DAILY', 'EACH_VISIT', 'WEEKLY'))
);

CREATE TABLE proposal_items (
    revision_id uuid NOT NULL REFERENCES proposal_revisions (id),
    product_id uuid NOT NULL REFERENCES products (id),
    action text NOT NULL,
    pricing_model text NOT NULL,
    retail_price bigint NOT NULL,
    commission_bps integer NOT NULL DEFAULT 0,
    deposit_price bigint NOT NULL DEFAULT 0,
    PRIMARY KEY (revision_id, product_id),
    CONSTRAINT proposal_items_action_known CHECK (action IN ('ADD', 'UPDATE', 'STOP')),
    CONSTRAINT proposal_items_model_known CHECK (pricing_model IN ('FIXED_DEPOSIT', 'COMMISSION')),
    CONSTRAINT proposal_items_prices_range CHECK (
        retail_price BETWEEN 0 AND 100000000 AND deposit_price BETWEEN 0 AND retail_price
    ),
    CONSTRAINT proposal_items_commission_range CHECK (commission_bps BETWEEN 0 AND 5000)
);

CREATE TABLE agreements (
    id uuid PRIMARY KEY,
    partnership_id uuid NOT NULL REFERENCES partnerships (id),
    version integer NOT NULL,
    status text NOT NULL DEFAULT 'ACTIVE',
    settlement_schedule text NOT NULL,
    payment_terms_days integer NOT NULL,
    count_rhythm text NOT NULL,
    notes text NOT NULL DEFAULT '',
    source_proposal_id uuid NOT NULL REFERENCES proposals (id),
    proposed_by uuid NOT NULL REFERENCES users (id),
    accepted_by uuid NOT NULL REFERENCES users (id),
    accepted_at timestamptz NOT NULL,
    effective_from timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT agreements_version_key UNIQUE (partnership_id, version),
    CONSTRAINT agreements_status_known CHECK (status IN ('ACTIVE', 'SUPERSEDED')),
    CONSTRAINT agreements_version_positive CHECK (version > 0),
    CONSTRAINT agreements_schedule_known CHECK (settlement_schedule IN ('ON_COUNT', 'WEEKLY', 'MONTHLY')),
    CONSTRAINT agreements_payment_terms_range CHECK (payment_terms_days BETWEEN 0 AND 30),
    CONSTRAINT agreements_rhythm_known CHECK (count_rhythm IN ('DAILY', 'EACH_VISIT', 'WEEKLY'))
);

CREATE UNIQUE INDEX agreements_active_key ON agreements (partnership_id) WHERE status = 'ACTIVE';

CREATE TABLE agreement_items (
    agreement_id uuid NOT NULL REFERENCES agreements (id),
    product_id uuid NOT NULL REFERENCES products (id),
    pricing_model text NOT NULL,
    retail_price bigint NOT NULL,
    commission_bps integer NOT NULL DEFAULT 0,
    deposit_price bigint NOT NULL DEFAULT 0,
    is_active boolean NOT NULL DEFAULT true,
    PRIMARY KEY (agreement_id, product_id),
    CONSTRAINT agreement_items_model_known CHECK (pricing_model IN ('FIXED_DEPOSIT', 'COMMISSION')),
    CONSTRAINT agreement_items_prices_range CHECK (
        retail_price BETWEEN 0 AND 100000000 AND deposit_price BETWEEN 0 AND retail_price
    ),
    CONSTRAINT agreement_items_commission_range CHECK (commission_bps BETWEEN 0 AND 5000)
);

-- +goose Down
DROP TABLE agreement_items;
DROP TABLE agreements;
DROP TABLE proposal_items;
DROP TABLE proposal_revisions;
DROP TABLE proposals;
DROP TABLE partnerships;
DROP TABLE document_sequences;
