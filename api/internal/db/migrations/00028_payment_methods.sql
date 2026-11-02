-- +goose Up
ALTER TABLE organizations
    ADD COLUMN payment_methods text[] NOT NULL DEFAULT ARRAY['CASH']::text[],
    ADD COLUMN bank_name text NOT NULL DEFAULT '',
    ADD COLUMN bank_account_number text NOT NULL DEFAULT '',
    ADD COLUMN bank_account_holder text NOT NULL DEFAULT '',
    ADD CONSTRAINT organizations_payment_methods_known CHECK (
        cardinality(payment_methods) BETWEEN 1 AND 3
        AND payment_methods <@ ARRAY['CASH', 'TRANSFER', 'QRIS']::text[]
    );

UPDATE organizations
SET payment_methods = array_remove(ARRAY[
    'CASH',
    CASE WHEN payment_info <> '' THEN 'TRANSFER' END,
    CASE WHEN payment_qr_file_id IS NOT NULL THEN 'QRIS' END
], NULL)
WHERE type = 'CONSIGNOR';

ALTER TABLE files DROP CONSTRAINT files_content_type_image;
ALTER TABLE files ADD CONSTRAINT files_content_type_known CHECK (
    content_type IN ('image/jpeg', 'image/png')
    OR (content_type = 'application/pdf' AND purpose = 'PAYMENT_PROOF')
);
ALTER TABLE files DROP CONSTRAINT files_dimension_limit;
ALTER TABLE files ADD CONSTRAINT files_dimension_limit CHECK (
    (content_type = 'application/pdf' AND width = 0 AND height = 0)
    OR (width BETWEEN 1 AND 4096 AND height BETWEEN 1 AND 4096)
);

-- +goose Down
ALTER TABLE files DROP CONSTRAINT files_dimension_limit;
ALTER TABLE files ADD CONSTRAINT files_dimension_limit
    CHECK (width BETWEEN 1 AND 4096 AND height BETWEEN 1 AND 4096);
ALTER TABLE files DROP CONSTRAINT files_content_type_known;
ALTER TABLE files ADD CONSTRAINT files_content_type_image CHECK (content_type IN ('image/jpeg', 'image/png'));

ALTER TABLE organizations
    DROP CONSTRAINT organizations_payment_methods_known,
    DROP COLUMN bank_account_holder,
    DROP COLUMN bank_account_number,
    DROP COLUMN bank_name,
    DROP COLUMN payment_methods;
