-- name: GetSettlementLedger :one
SELECT
    p.id, p.consignor_org_id, p.store_org_id, s.name AS store_name,
    COALESCE(a.settlement_schedule, '')::text AS settlement_schedule,
    COALESCE(a.payment_terms_days, 0)::bigint AS payment_terms_days,
    COALESCE(
        (SELECT max(st.period_end) FROM settlements AS st WHERE st.partnership_id = p.id),
        '0001-01-01T00:00:00Z'::timestamptz
    )::timestamptz AS settled_until
FROM partnerships AS p
JOIN organizations AS s ON s.id = p.store_org_id
LEFT JOIN agreements AS a ON a.partnership_id = p.id AND a.status = 'ACTIVE'
WHERE p.id = @partnership_id::uuid;

-- name: ListUnsettledAllocations :many
SELECT
    a.id, a.product_id, a.occurred_at, a.qty, a.gross_amount, a.store_share_amount, a.consignor_share_amount,
    i.pricing_model, i.retail_price, i.commission_bps, i.deposit_price
FROM sale_allocations AS a
JOIN drop_items AS i ON i.id = a.drop_item_id
WHERE a.partnership_id = @partnership_id::uuid AND a.settlement_id IS NULL
ORDER BY a.occurred_at, a.id;

-- name: CreateSettlement :exec
INSERT INTO settlements (
    id, number, partnership_id, consignor_org_id, store_org_id, schedule, period_start, period_end, status,
    sold_qty, gross_amount, store_share_amount, consignor_share_amount, due_date, issued_at, created_at, updated_at
) VALUES (
    @id::uuid, @number::text, @partnership_id::uuid, @consignor_org_id::uuid, @store_org_id::uuid,
    @schedule::text, @period_start::timestamptz, @period_end::timestamptz, 'ISSUED', @sold_qty::bigint,
    @gross_amount::bigint, @store_share_amount::bigint, @consignor_share_amount::bigint,
    CAST(@due_date::text AS date),
    @issued_at::timestamptz, @issued_at::timestamptz, @issued_at::timestamptz
);

-- name: LinkSettlementAllocations :execrows
UPDATE sale_allocations
SET settlement_id = @settlement_id::uuid
WHERE id = ANY(@ids::uuid[]) AND partnership_id = @partnership_id::uuid AND settlement_id IS NULL;

-- name: GetSettlement :one
SELECT
    st.id, st.number, st.partnership_id, st.consignor_org_id, c.name AS consignor_name, st.store_org_id,
    s.name AS store_name, st.schedule, st.period_start, st.period_end, st.status, st.sold_qty, st.gross_amount,
    st.store_share_amount, st.consignor_share_amount, to_char(st.due_date, 'YYYY-MM-DD')::text AS due_date,
    st.issued_at, COALESCE(st.paid_at, st.issued_at)::timestamptz AS paid_at,
    (st.consignor_org_id = @organization_id::uuid OR pa.status = 'ACTIVE')::boolean AS shows_payee,
    c.payment_info, COALESCE(c.payment_qr_file_id::text, '')::text AS payment_qr_file_id
FROM settlements AS st
JOIN partnerships AS pa ON pa.id = st.partnership_id
JOIN organizations AS c ON c.id = st.consignor_org_id
JOIN organizations AS s ON s.id = st.store_org_id
WHERE st.id = @id::uuid
    AND (st.consignor_org_id = @organization_id::uuid OR st.store_org_id = @organization_id::uuid);

-- name: ListSettlementAllocations :many
SELECT
    a.id, a.product_id, pr.name AS product_name, pr.sku, pr.unit, a.occurred_at, a.qty, a.gross_amount,
    a.store_share_amount, a.consignor_share_amount, i.pricing_model, i.retail_price, i.commission_bps,
    i.deposit_price, c.id AS count_id, c.number AS count_number, c.counted_at
FROM sale_allocations AS a
JOIN drop_items AS i ON i.id = a.drop_item_id
JOIN products AS pr ON pr.id = a.product_id
JOIN count_lines AS l ON l.id = a.count_line_id
JOIN counts AS c ON c.id = l.count_id
WHERE a.settlement_id = @settlement_id::uuid
ORDER BY a.occurred_at, a.id;

-- name: ListSettlements :many
SELECT
    st.id, st.number, st.partnership_id, st.status, st.schedule, st.period_start, st.period_end,
    to_char(st.due_date, 'YYYY-MM-DD')::text AS due_date, st.issued_at, st.sold_qty, st.consignor_share_amount,
    o.id AS counterpart_id, o.name AS counterpart_name, o.type AS counterpart_type
FROM settlements AS st
JOIN organizations AS o
    ON o.id = CASE WHEN st.store_org_id = @organization_id::uuid THEN st.consignor_org_id ELSE st.store_org_id END
WHERE (st.consignor_org_id = @organization_id::uuid OR st.store_org_id = @organization_id::uuid)
    AND (@partnership_id::text = '' OR st.partnership_id::text = @partnership_id::text)
    AND (@status::text = '' OR st.status = @status::text)
    AND st.due_date BETWEEN CAST(@due_from::text AS date) AND CAST(@due_to::text AS date)
    AND (@from_start::boolean OR (st.issued_at, st.id) < (@after_at::timestamptz, @after_id::uuid))
ORDER BY st.issued_at DESC, st.id DESC
LIMIT @row_limit::bigint;

-- name: GetRunningSales :one
SELECT
    COUNT(a.id)::bigint AS allocation_count, COALESCE(SUM(a.qty), 0)::bigint AS sold_qty,
    COALESCE(SUM(a.gross_amount), 0)::bigint AS gross_amount,
    COALESCE(SUM(a.store_share_amount), 0)::bigint AS store_share_amount,
    COALESCE(SUM(a.consignor_share_amount), 0)::bigint AS consignor_share_amount,
    COALESCE(MIN(a.occurred_at), '0001-01-01T00:00:00Z'::timestamptz)::timestamptz AS since
FROM partnerships AS p
LEFT JOIN sale_allocations AS a ON a.partnership_id = p.id AND a.settlement_id IS NULL
WHERE p.id = @partnership_id::uuid
    AND (p.consignor_org_id = @organization_id::uuid OR p.store_org_id = @organization_id::uuid)
GROUP BY p.id;

-- name: ListPartnershipsDueForSettlement :many
SELECT DISTINCT a.partnership_id
FROM sale_allocations AS a
JOIN agreements AS ag ON ag.partnership_id = a.partnership_id AND ag.status = 'ACTIVE'
WHERE a.settlement_id IS NULL AND a.occurred_at < @before::timestamptz
    AND ag.settlement_schedule = @schedule::text
ORDER BY a.partnership_id;

-- name: LockSettlementPartnership :one
SELECT id
FROM partnerships
WHERE id = @partnership_id::uuid
FOR UPDATE;

-- name: FindPaymentByKey :one
SELECT id, settlement_id, request_fingerprint
FROM payments
WHERE recorded_by_org_id = @organization_id::uuid AND idempotency_key = @idempotency_key::uuid;

-- name: LockPayable :one
SELECT
    st.id, st.number, st.partnership_id, st.consignor_org_id, st.store_org_id, s.name AS store_name, st.status,
    st.consignor_share_amount,
    EXISTS (
        SELECT 1 FROM payments AS p
        WHERE p.settlement_id = st.id AND p.status IN ('PENDING_CONFIRMATION', 'CONFIRMED')
    )::boolean AS has_open_payment
FROM settlements AS st
JOIN organizations AS s ON s.id = st.store_org_id
WHERE st.id = @id::uuid
    AND (st.consignor_org_id = @organization_id::uuid OR st.store_org_id = @organization_id::uuid)
FOR UPDATE OF st;

-- name: GetProofFile :one
SELECT sha256
FROM files
WHERE id = @id::uuid AND organization_id = @organization_id::uuid AND purpose = 'PAYMENT_PROOF';

-- name: CreatePayment :execrows
INSERT INTO payments (
    id, settlement_id, partnership_id, amount, method, paid_on, proof_file_id, proof_sha256, status,
    recorded_by, recorded_by_org_id, reviewed_by, reviewed_at, idempotency_key, request_fingerprint, created_at,
    updated_at
) VALUES (
    @id::uuid, @settlement_id::uuid, @partnership_id::uuid, @amount::bigint, @method::text,
    CAST(@paid_on::text AS date), NULLIF(@proof_file_id::text, '')::uuid,
    CASE WHEN @proof_file_id::text = '' THEN NULL ELSE @proof_sha256::bytea END, @status::text,
    @recorded_by::uuid, @recorded_by_org_id::uuid,
    CASE WHEN @status::text = 'CONFIRMED' THEN @recorded_by::uuid END,
    CASE WHEN @status::text = 'CONFIRMED' THEN @recorded_at::timestamptz END,
    @idempotency_key::uuid, @request_fingerprint::bytea, @recorded_at::timestamptz, @recorded_at::timestamptz
)
ON CONFLICT (recorded_by_org_id, idempotency_key) DO NOTHING;

-- name: LockPayment :one
SELECT
    p.id, p.settlement_id, st.number AS settlement_number, p.partnership_id, st.consignor_org_id, st.store_org_id,
    s.name AS store_name, p.amount, p.method, to_char(p.paid_on, 'YYYY-MM-DD')::text AS paid_on,
    COALESCE(p.proof_file_id::text, '')::text AS proof_file_id,
    COALESCE(p.proof_sha256, ''::bytea)::bytea AS proof_sha256,
    p.status, p.recorded_by, p.recorded_by_org_id, p.created_at
FROM payments AS p
JOIN settlements AS st ON st.id = p.settlement_id
JOIN organizations AS s ON s.id = st.store_org_id
WHERE p.id = @id::uuid
    AND (st.consignor_org_id = @organization_id::uuid OR st.store_org_id = @organization_id::uuid)
FOR UPDATE OF p, st;

-- name: ReviewPayment :execrows
UPDATE payments
SET status = @status::text, rejection_reason = @rejection_reason::text, reviewed_by = @reviewed_by::uuid,
    reviewed_at = @reviewed_at::timestamptz, updated_at = @reviewed_at::timestamptz
WHERE id = @id::uuid AND status = 'PENDING_CONFIRMATION';

-- name: MarkSettlementPaid :execrows
UPDATE settlements
SET status = 'PAID', paid_at = @paid_at::timestamptz, updated_at = @paid_at::timestamptz
WHERE id = @id::uuid AND status = 'ISSUED';

-- name: ListSettlementPayments :many
SELECT
    p.id, p.amount, p.method, to_char(p.paid_on, 'YYYY-MM-DD')::text AS paid_on,
    COALESCE(p.proof_file_id::text, '')::text AS proof_file_id,
    COALESCE(p.proof_sha256, ''::bytea)::bytea AS proof_sha256,
    p.status, p.rejection_reason, p.recorded_by, p.recorded_by_org_id, ru.full_name AS recorded_by_name,
    COALESCE(vu.full_name, '')::text AS reviewed_by_name,
    COALESCE(p.reviewed_at, p.created_at)::timestamptz AS reviewed_at, p.created_at
FROM payments AS p
JOIN users AS ru ON ru.id = p.recorded_by
LEFT JOIN users AS vu ON vu.id = p.reviewed_by
WHERE p.settlement_id = @settlement_id::uuid
ORDER BY p.created_at, p.id;
