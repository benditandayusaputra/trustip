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
    is_confirmed, d.auto_confirmed
FROM drops AS d
JOIN partnerships AS p ON p.id = d.partnership_id
JOIN agreements AS a ON a.id = d.agreement_id
JOIN organizations AS s ON s.id = d.store_org_id
JOIN organizations AS c ON c.id = p.consignor_org_id
JOIN users AS ru ON ru.id = d.recorded_by
LEFT JOIN users AS cu ON cu.id = d.confirmed_by
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
