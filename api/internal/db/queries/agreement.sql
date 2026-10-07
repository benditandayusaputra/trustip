-- name: CreateAgreement :exec
INSERT INTO agreements (
    id, partnership_id, version, status, settlement_schedule, payment_terms_days, count_rhythm, notes,
    source_proposal_id, proposed_by, accepted_by, accepted_at, effective_from, created_at
) VALUES (
    @id::uuid, @partnership_id::uuid, @version::bigint, 'ACTIVE', @settlement_schedule::text,
    @payment_terms_days::bigint, @count_rhythm::text, @notes::text, @source_proposal_id::uuid, @proposed_by::uuid,
    @accepted_by::uuid, @accepted_at::timestamptz, @accepted_at::timestamptz, @accepted_at::timestamptz
);

-- name: AddAgreementItems :exec
INSERT INTO agreement_items (agreement_id, product_id, pricing_model, retail_price, commission_bps, deposit_price)
SELECT @agreement_id::uuid, item.product_id, item.pricing_model, item.retail_price, item.commission_bps,
    item.deposit_price
FROM (
    SELECT
        unnest(@product_ids::uuid[]) AS product_id, unnest(@pricing_models::text[]) AS pricing_model,
        unnest(@retail_prices::bigint[]) AS retail_price, unnest(@commission_bps::bigint[]) AS commission_bps,
        unnest(@deposit_prices::bigint[]) AS deposit_price
) AS item;
