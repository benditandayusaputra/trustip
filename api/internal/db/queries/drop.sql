-- name: GetDropPartnership :one
SELECT
    p.id, p.status, p.store_org_id, s.name AS store_name, p.consignor_org_id,
    COALESCE(a.id::text, '')::text AS agreement_id, COALESCE(a.version, 0)::bigint AS agreement_version
FROM partnerships AS p
JOIN organizations AS s ON s.id = p.store_org_id
LEFT JOIN agreements AS a ON a.partnership_id = p.id AND a.status = 'ACTIVE'
WHERE p.id = @partnership_id::uuid AND p.store_org_id = @store_org_id::uuid;

-- name: ListDropProducts :many
SELECT
    i.product_id, pr.shelf_life_days, i.pricing_model, i.retail_price, i.commission_bps, i.deposit_price
FROM agreement_items AS i
JOIN products AS pr ON pr.id = i.product_id
WHERE i.agreement_id = @agreement_id::uuid AND i.is_active;

-- name: FindDropByIdempotencyKey :one
SELECT id, request_fingerprint
FROM drops
WHERE store_org_id = @store_org_id::uuid AND idempotency_key = @idempotency_key::uuid;

-- name: CreateDrop :execrows
INSERT INTO drops (
    id, number, partnership_id, agreement_id, store_org_id, initiated_by, status, occurred_at, recorded_by,
    idempotency_key, request_fingerprint, client_recorded_at, received_at, created_at, updated_at
) VALUES (
    @id::uuid, @number::text, @partnership_id::uuid, @agreement_id::uuid, @store_org_id::uuid, 'STORE',
    'RECORDED', @occurred_at::timestamptz, @recorded_by::uuid, @idempotency_key::uuid,
    @request_fingerprint::bytea, @received_at::timestamptz, @received_at::timestamptz, @received_at::timestamptz,
    @received_at::timestamptz
)
ON CONFLICT (store_org_id, idempotency_key) DO NOTHING;

-- name: AddDropItems :exec
INSERT INTO drop_items (
    id, drop_id, line_no, product_id, batch_code, expiry_date, qty_declared, qty_received, pricing_model,
    retail_price, commission_bps, deposit_price
)
SELECT
    item.id, @drop_id::uuid, item.line_no, item.product_id, item.batch_code, item.expiry_date, item.qty, item.qty,
    item.pricing_model, item.retail_price, item.commission_bps, item.deposit_price
FROM (
    SELECT
        unnest(@ids::uuid[]) AS id, unnest(@line_numbers::bigint[]) AS line_no,
        unnest(@product_ids::uuid[]) AS product_id, unnest(@batch_codes::text[]) AS batch_code,
        NULLIF(unnest(@expiry_dates::text[]), '')::date AS expiry_date, unnest(@quantities::bigint[]) AS qty,
        unnest(@pricing_models::text[]) AS pricing_model, unnest(@retail_prices::bigint[]) AS retail_price,
        unnest(@commission_bps::bigint[]) AS commission_bps, unnest(@deposit_prices::bigint[]) AS deposit_price
) AS item;

-- name: AddStockMovements :exec
INSERT INTO stock_movements (
    id, partnership_id, product_id, drop_item_id, type, qty, occurred_at, source_type, source_id, source_number,
    created_by
)
SELECT
    movement.id, @partnership_id::uuid, movement.product_id, movement.drop_item_id, movement.type, movement.qty,
    @occurred_at::timestamptz, @source_type::text, @source_id::uuid, @source_number::text, @created_by::uuid
FROM (
    SELECT
        unnest(@ids::uuid[]) AS id, unnest(@product_ids::uuid[]) AS product_id,
        unnest(@drop_item_ids::uuid[]) AS drop_item_id, unnest(@types::text[]) AS type,
        unnest(@quantities::bigint[]) AS qty
) AS movement;

-- name: GetDrop :one
SELECT
    d.id, d.number, d.partnership_id, d.agreement_id, a.version AS agreement_version, d.store_org_id,
    s.name AS store_name, p.consignor_org_id, c.name AS consignor_name, d.status, d.occurred_at, d.received_at,
    d.recorded_offline, d.recorded_by, ru.full_name AS recorded_by_name,
    COALESCE(cu.full_name, '')::text AS confirmed_by_name,
    COALESCE(d.confirmed_at, d.received_at)::timestamptz AS confirmed_at, (d.confirmed_at IS NOT NULL)::boolean AS
    is_confirmed, d.auto_confirmed, COALESCE(xu.full_name, '')::text AS cancelled_by_name,
    COALESCE(d.cancelled_at, d.received_at)::timestamptz AS cancelled_at
FROM drops AS d
JOIN partnerships AS p ON p.id = d.partnership_id
JOIN agreements AS a ON a.id = d.agreement_id
JOIN organizations AS s ON s.id = d.store_org_id
JOIN organizations AS c ON c.id = p.consignor_org_id
JOIN users AS ru ON ru.id = d.recorded_by
LEFT JOIN users AS cu ON cu.id = d.confirmed_by
LEFT JOIN users AS xu ON xu.id = d.cancelled_by
WHERE d.id = @id::uuid AND (p.store_org_id = @organization_id::uuid OR p.consignor_org_id = @organization_id::uuid);

-- name: ListDropItems :many
SELECT
    i.id, i.product_id, pr.name AS product_name, pr.sku, pr.unit, i.batch_code,
    COALESCE(to_char(i.expiry_date, 'YYYY-MM-DD'), '')::text AS expiry_date, i.qty_received, i.pricing_model,
    i.retail_price, i.commission_bps, i.deposit_price
FROM drop_items AS i
JOIN products AS pr ON pr.id = i.product_id
WHERE i.drop_id = @drop_id::uuid
ORDER BY i.line_no;

-- name: SumProductStock :one
SELECT COALESCE(sum(qty), 0)::bigint AS balance
FROM stock_movements
WHERE partnership_id = @partnership_id::uuid AND product_id = @product_id::uuid;

-- name: LockDrop :one
SELECT d.id
FROM drops AS d
JOIN partnerships AS p ON p.id = d.partnership_id
WHERE d.id = @id::uuid AND (p.store_org_id = @organization_id::uuid OR p.consignor_org_id = @organization_id::uuid)
FOR UPDATE OF d, p;

-- name: CountDropOutflows :one
SELECT count(*)::bigint AS outflows
FROM stock_movements AS m
JOIN drop_items AS i ON i.id = m.drop_item_id
WHERE i.drop_id = @drop_id::uuid AND m.type <> 'RECEIVE';

-- name: ConfirmDrop :execrows
UPDATE drops
SET status = 'CONFIRMED', confirmed_by = @confirmed_by::uuid, confirmed_at = @confirmed_at::timestamptz,
    auto_confirmed = false, updated_at = @confirmed_at::timestamptz
WHERE id = @id::uuid AND status = 'RECORDED';

-- name: CancelDrop :execrows
UPDATE drops
SET status = 'CANCELLED', cancelled_by = @cancelled_by::uuid, cancelled_at = @cancelled_at::timestamptz,
    updated_at = @cancelled_at::timestamptz
WHERE id = @id::uuid AND status = 'RECORDED';

-- name: ListDrops :many
SELECT
    d.id, d.number, d.partnership_id, d.occurred_at, d.recorded_offline,
    (CASE
        WHEN d.status = 'RECORDED' AND d.received_at <= @confirm_cutoff::timestamptz THEN 'CONFIRMED'
        ELSE d.status
    END)::text AS status,
    o.id AS counterpart_id, o.name AS counterpart_name, o.type AS counterpart_type,
    COALESCE(lines.total_qty, 0)::bigint AS total_qty, COALESCE(lines.line_count, 0)::bigint AS line_count
FROM drops AS d
JOIN partnerships AS p ON p.id = d.partnership_id
JOIN organizations AS o
    ON o.id = CASE WHEN p.store_org_id = @organization_id::uuid THEN p.consignor_org_id ELSE p.store_org_id END
LEFT JOIN LATERAL (
    SELECT sum(i.qty_received) AS total_qty, count(*) AS line_count
    FROM drop_items AS i
    WHERE i.drop_id = d.id
) AS lines ON true
WHERE (p.store_org_id = @organization_id::uuid OR p.consignor_org_id = @organization_id::uuid)
    AND (@partnership_id::text = '' OR d.partnership_id::text = @partnership_id::text)
    AND d.occurred_at >= @from_at::timestamptz AND d.occurred_at < @to_at::timestamptz
    AND (
        @status::text = ''
        OR (CASE
            WHEN d.status = 'RECORDED' AND d.received_at <= @confirm_cutoff::timestamptz THEN 'CONFIRMED'
            ELSE d.status
        END) = @status::text
    )
    AND (@from_start::boolean OR (d.occurred_at, d.id) < (@after_at::timestamptz, @after_id::uuid))
ORDER BY d.occurred_at DESC, d.id DESC
LIMIT @row_limit::bigint;

-- name: GetLastCountAt :one
SELECT
    (max(counted_at) IS NOT NULL)::boolean AS has_count,
    COALESCE(max(counted_at), '0001-01-01T00:00:00Z'::timestamptz)::timestamptz AS last_at
FROM counts
WHERE partnership_id = @partnership_id::uuid;

-- name: ConfirmUnansweredDrops :execrows
UPDATE drops
SET status = 'CONFIRMED', auto_confirmed = true,
    confirmed_at = received_at + make_interval(0, 0, 0, 0, 0, 0, @window_seconds::double precision), updated_at = @now::timestamptz
WHERE status = 'RECORDED' AND received_at <= @cutoff::timestamptz;
