-- name: ListPartnershipBatches :many
SELECT
    b.drop_item_id, b.product_id, pr.name AS product_name, pr.sku, pr.unit,
    COALESCE(pr.shelf_life_days, 0)::bigint AS shelf_life_days, b.batch_code,
    COALESCE(to_char(b.expiry_date, 'YYYY-MM-DD'), '')::text AS expiry_date, b.received_at, b.balance
FROM batch_balances AS b
JOIN partnerships AS p ON p.id = b.partnership_id
JOIN products AS pr ON pr.id = b.product_id
WHERE b.partnership_id = @partnership_id::uuid AND b.balance > 0
    AND (p.store_org_id = @organization_id::uuid OR p.consignor_org_id = @organization_id::uuid)
ORDER BY pr.name, pr.id, b.expiry_date NULLS LAST, b.received_at, b.drop_item_id;

-- name: ListOrganizationBatches :many
SELECT
    b.drop_item_id, b.partnership_id, b.product_id, pr.name AS product_name, pr.sku, pr.unit,
    COALESCE(pr.shelf_life_days, 0)::bigint AS shelf_life_days,
    COALESCE(to_char(b.expiry_date, 'YYYY-MM-DD'), '')::text AS expiry_date, b.received_at, b.balance,
    o.id AS counterpart_id, o.name AS counterpart_name, o.type AS counterpart_type
FROM batch_balances AS b
JOIN partnerships AS p ON p.id = b.partnership_id
JOIN products AS pr ON pr.id = b.product_id
JOIN organizations AS o
    ON o.id = CASE WHEN p.store_org_id = @organization_id::uuid THEN p.consignor_org_id ELSE p.store_org_id END
WHERE b.balance > 0 AND (p.store_org_id = @organization_id::uuid OR p.consignor_org_id = @organization_id::uuid)
ORDER BY pr.name, pr.id, o.name, b.partnership_id;

-- name: ListStockCard :many
WITH card AS (
    SELECT
        m.id, m.type, m.qty, m.occurred_at, i.batch_code, m.source_type, m.source_id, m.source_number,
        sum(m.qty) OVER (ORDER BY m.occurred_at, m.id)::bigint AS balance_after
    FROM stock_movements AS m
    JOIN drop_items AS i ON i.id = m.drop_item_id
    JOIN partnerships AS p ON p.id = m.partnership_id
    WHERE m.partnership_id = @partnership_id::uuid AND m.product_id = @product_id::uuid
        AND (p.store_org_id = @organization_id::uuid OR p.consignor_org_id = @organization_id::uuid)
)
SELECT id, type, qty, occurred_at, batch_code, source_type, source_id, source_number, balance_after
FROM card
WHERE @from_start::boolean OR (occurred_at, id) < (@after_at::timestamptz, @after_id::uuid)
ORDER BY occurred_at DESC, id DESC
LIMIT @row_limit::bigint;

-- name: GetStockCardProduct :one
SELECT pr.id, pr.name, pr.sku, pr.unit
FROM partnerships AS p
JOIN products AS pr ON pr.consignor_org_id = p.consignor_org_id
WHERE p.id = @partnership_id::uuid AND pr.id = @product_id::uuid
    AND (p.store_org_id = @organization_id::uuid OR p.consignor_org_id = @organization_id::uuid);

-- name: FindStockPartnership :one
SELECT p.id
FROM partnerships AS p
WHERE p.id = @partnership_id::uuid
    AND (p.store_org_id = @organization_id::uuid OR p.consignor_org_id = @organization_id::uuid);
