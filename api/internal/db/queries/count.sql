-- name: LockPartnershipRow :one
SELECT id
FROM partnerships
WHERE id = @partnership_id::uuid AND store_org_id = @store_org_id::uuid
FOR UPDATE;

-- name: GetCountPartnership :one
SELECT
    p.id, p.status, p.store_org_id, s.name AS store_name, p.consignor_org_id, c.name AS consignor_name,
    COALESCE(a.version, 0)::bigint AS agreement_version
FROM partnerships AS p
JOIN organizations AS s ON s.id = p.store_org_id
JOIN organizations AS c ON c.id = p.consignor_org_id
LEFT JOIN agreements AS a ON a.partnership_id = p.id AND a.status = 'ACTIVE'
WHERE p.id = @partnership_id::uuid AND p.store_org_id = @store_org_id::uuid;

-- name: ListCountBatches :many
SELECT
    b.drop_item_id, b.product_id, pr.name AS product_name, pr.sku, pr.unit,
    COALESCE(to_char(b.expiry_date, 'YYYY-MM-DD'), '')::text AS expiry_date, b.received_at, b.balance,
    i.pricing_model, i.retail_price, i.commission_bps, i.deposit_price
FROM batch_balances AS b
JOIN drop_items AS i ON i.id = b.drop_item_id
JOIN products AS pr ON pr.id = b.product_id
WHERE b.partnership_id = @partnership_id::uuid AND b.balance > 0
ORDER BY pr.name, pr.id, b.expiry_date NULLS LAST, b.received_at, b.drop_item_id;

-- name: GetLastDocumentAt :one
SELECT
    (latest.at IS NOT NULL)::boolean AS has_document,
    COALESCE(latest.at, '0001-01-01T00:00:00Z'::timestamptz)::timestamptz AS last_at
FROM (
    SELECT max(at) AS at
    FROM (
        SELECT occurred_at AS at FROM drops WHERE partnership_id = @partnership_id::uuid AND status <> 'CANCELLED'
        UNION ALL
        SELECT counted_at AS at FROM counts WHERE partnership_id = @partnership_id::uuid
    ) AS documents
) AS latest;

-- name: FindCountByIdempotencyKey :one
SELECT id, request_fingerprint
FROM counts
WHERE store_org_id = @store_org_id::uuid AND idempotency_key = @idempotency_key::uuid;

-- name: CreateCount :execrows
INSERT INTO counts (
    id, number, partnership_id, store_org_id, occasion, counted_at, status, note, sold_qty, returned_qty,
    gross_amount, store_share_amount, consignor_share_amount, recorded_by, idempotency_key, request_fingerprint,
    client_recorded_at, received_at, created_at, updated_at
) VALUES (
    @id::uuid, @number::text, @partnership_id::uuid, @store_org_id::uuid, @occasion::text,
    @counted_at::timestamptz, 'RECORDED', @note::text, @sold_qty::bigint, @returned_qty::bigint,
    @gross_amount::bigint, @store_share_amount::bigint, @consignor_share_amount::bigint, @recorded_by::uuid,
    @idempotency_key::uuid, @request_fingerprint::bytea, @received_at::timestamptz, @received_at::timestamptz,
    @received_at::timestamptz, @received_at::timestamptz
)
ON CONFLICT (store_org_id, idempotency_key) DO NOTHING;

-- name: AddCountLines :exec
INSERT INTO count_lines (
    id, count_id, line_no, product_id, recorded_qty, counted_qty, sold_qty, returned_qty, kept_qty, return_reason
)
SELECT
    line.id, @count_id::uuid, line.line_no, line.product_id, line.recorded_qty, line.counted_qty, line.sold_qty,
    line.returned_qty, line.kept_qty, line.return_reason
FROM (
    SELECT
        unnest(@ids::uuid[]) AS id, unnest(@line_numbers::bigint[]) AS line_no,
        unnest(@product_ids::uuid[]) AS product_id, unnest(@recorded::bigint[]) AS recorded_qty,
        unnest(@counted::bigint[]) AS counted_qty, unnest(@sold::bigint[]) AS sold_qty,
        unnest(@returned::bigint[]) AS returned_qty, unnest(@kept::bigint[]) AS kept_qty,
        unnest(@reasons::text[]) AS return_reason
) AS line;

-- name: AddSaleAllocations :exec
INSERT INTO sale_allocations (
    id, count_line_id, drop_item_id, partnership_id, product_id, qty, gross_amount, store_share_amount,
    consignor_share_amount, occurred_at
)
SELECT
    sale.id, sale.count_line_id, sale.drop_item_id, @partnership_id::uuid, sale.product_id, sale.qty,
    sale.gross_amount, sale.store_share_amount, sale.consignor_share_amount, @occurred_at::timestamptz
FROM (
    SELECT
        unnest(@ids::uuid[]) AS id, unnest(@count_line_ids::uuid[]) AS count_line_id,
        unnest(@drop_item_ids::uuid[]) AS drop_item_id, unnest(@product_ids::uuid[]) AS product_id,
        unnest(@quantities::bigint[]) AS qty, unnest(@gross_amounts::bigint[]) AS gross_amount,
        unnest(@store_shares::bigint[]) AS store_share_amount,
        unnest(@consignor_shares::bigint[]) AS consignor_share_amount
) AS sale;

-- name: AddReturnAllocations :exec
INSERT INTO return_allocations (id, count_line_id, drop_item_id, qty)
SELECT taken.id, taken.count_line_id, taken.drop_item_id, taken.qty
FROM (
    SELECT
        unnest(@ids::uuid[]) AS id, unnest(@count_line_ids::uuid[]) AS count_line_id,
        unnest(@drop_item_ids::uuid[]) AS drop_item_id, unnest(@quantities::bigint[]) AS qty
) AS taken;

-- name: GetCount :one
SELECT
    c.id, c.number, c.partnership_id, c.store_org_id, s.name AS store_name, p.consignor_org_id,
    o.name AS consignor_name, c.occasion, c.counted_at, c.received_at, c.status, c.note, c.sold_qty,
    c.returned_qty, c.gross_amount, c.store_share_amount, c.consignor_share_amount, c.recorded_by,
    ru.full_name AS recorded_by_name, COALESCE(cu.full_name, '')::text AS confirmed_by_name,
    COALESCE(c.confirmed_at, c.received_at)::timestamptz AS confirmed_at,
    (c.confirmed_at IS NOT NULL)::boolean AS is_confirmed, c.auto_confirmed, c.recorded_offline
FROM counts AS c
JOIN partnerships AS p ON p.id = c.partnership_id
JOIN organizations AS s ON s.id = c.store_org_id
JOIN organizations AS o ON o.id = p.consignor_org_id
JOIN users AS ru ON ru.id = c.recorded_by
LEFT JOIN users AS cu ON cu.id = c.confirmed_by
WHERE c.id = @id::uuid AND (p.store_org_id = @organization_id::uuid OR p.consignor_org_id = @organization_id::uuid);

-- name: ListCountLines :many
SELECT
    l.id, l.line_no, l.product_id, pr.name AS product_name, pr.sku, pr.unit, l.recorded_qty, l.counted_qty,
    l.sold_qty, l.returned_qty, l.kept_qty, l.return_reason
FROM count_lines AS l
JOIN products AS pr ON pr.id = l.product_id
WHERE l.count_id = @count_id::uuid
ORDER BY l.line_no;

-- name: ListCountSales :many
SELECT
    a.id, a.count_line_id, a.drop_item_id, a.qty, a.gross_amount, a.store_share_amount, a.consignor_share_amount
FROM sale_allocations AS a
JOIN count_lines AS l ON l.id = a.count_line_id
WHERE l.count_id = @count_id::uuid
ORDER BY l.line_no, a.id;

-- name: ListCountReturns :many
SELECT r.id, r.count_line_id, r.drop_item_id, r.qty
FROM return_allocations AS r
JOIN count_lines AS l ON l.id = r.count_line_id
WHERE l.count_id = @count_id::uuid
ORDER BY l.line_no, r.id;

-- name: LockCount :one
SELECT c.id
FROM counts AS c
JOIN partnerships AS p ON p.id = c.partnership_id
WHERE c.id = @id::uuid AND (p.store_org_id = @organization_id::uuid OR p.consignor_org_id = @organization_id::uuid)
FOR UPDATE OF c, p;

-- name: ConfirmCount :execrows
UPDATE counts
SET status = 'CONFIRMED', confirmed_by = @confirmed_by::uuid, confirmed_at = @confirmed_at::timestamptz,
    auto_confirmed = false, updated_at = @confirmed_at::timestamptz
WHERE id = @id::uuid AND status = 'RECORDED';

-- name: ListCounts :many
SELECT
    c.id, c.number, c.partnership_id, c.counted_at, c.recorded_offline,
    (CASE
        WHEN c.status = 'RECORDED' AND c.received_at <= @confirm_cutoff::timestamptz THEN 'CONFIRMED'
        ELSE c.status
    END)::text AS status,
    o.id AS counterpart_id, o.name AS counterpart_name, o.type AS counterpart_type,
    c.sold_qty, c.returned_qty, c.consignor_share_amount
FROM counts AS c
JOIN partnerships AS p ON p.id = c.partnership_id
JOIN organizations AS o
    ON o.id = CASE WHEN p.store_org_id = @organization_id::uuid THEN p.consignor_org_id ELSE p.store_org_id END
WHERE (p.store_org_id = @organization_id::uuid OR p.consignor_org_id = @organization_id::uuid)
    AND (@partnership_id::text = '' OR c.partnership_id::text = @partnership_id::text)
    AND c.counted_at >= @from_at::timestamptz AND c.counted_at < @to_at::timestamptz
    AND (@from_start::boolean OR (c.counted_at, c.id) < (@after_at::timestamptz, @after_id::uuid))
ORDER BY c.counted_at DESC, c.id DESC
LIMIT @row_limit::bigint;

-- name: GetSettledUntil :one
SELECT
    (latest.number IS NOT NULL)::boolean AS has_settlement, COALESCE(latest.number, '')::text AS number,
    COALESCE(latest.period_end, '0001-01-01T00:00:00Z'::timestamptz)::timestamptz AS period_end
FROM (SELECT 1) AS anchor
LEFT JOIN LATERAL (
    SELECT st.number, st.period_end
    FROM settlements AS st
    WHERE st.partnership_id = @partnership_id::uuid
    ORDER BY st.period_end DESC
    LIMIT 1
) AS latest ON true;

-- name: ListCountSettlements :many
SELECT DISTINCT st.id, st.number
FROM sale_allocations AS a
JOIN count_lines AS l ON l.id = a.count_line_id
JOIN settlements AS st ON st.id = a.settlement_id
WHERE l.count_id = @count_id::uuid
ORDER BY st.number;

-- name: ConfirmUnansweredCounts :execrows
UPDATE counts
SET status = 'CONFIRMED', auto_confirmed = true,
    confirmed_at = received_at + make_interval(0, 0, 0, 0, 0, 0, @window_seconds::double precision), updated_at = @now::timestamptz
WHERE status = 'RECORDED' AND received_at <= @cutoff::timestamptz;
