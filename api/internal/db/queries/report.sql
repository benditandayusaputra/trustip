-- name: SalesByPartner :many
WITH scoped AS (
    SELECT p.id, p.consignor_org_id, p.store_org_id
    FROM partnerships AS p
    WHERE p.store_org_id = @organization_id::uuid OR p.consignor_org_id = @organization_id::uuid
), facts AS (
    SELECT
        a.partnership_id, a.product_id, a.occurred_at, a.qty::bigint AS sold_qty, a.gross_amount,
        a.store_share_amount, a.consignor_share_amount, 0::bigint AS returned_qty, 0::bigint AS received_qty
    FROM sale_allocations AS a
    WHERE a.partnership_id IN (SELECT s.id FROM scoped AS s)
        AND a.occurred_at >= @from_at::timestamptz AND a.occurred_at < @to_at::timestamptz
    UNION ALL
    SELECT
        m.partnership_id, m.product_id, m.occurred_at, 0::bigint, 0::bigint, 0::bigint, 0::bigint,
        (CASE WHEN m.type = 'RETURN' THEN -m.qty ELSE 0 END)::bigint,
        (CASE WHEN m.source_type = 'DROP' THEN m.qty ELSE 0 END)::bigint
    FROM stock_movements AS m
    WHERE m.partnership_id IN (SELECT s.id FROM scoped AS s) AND m.type IN ('RETURN', 'RECEIVE', 'ADJUSTMENT')
        AND m.occurred_at >= @from_at::timestamptz AND m.occurred_at < @to_at::timestamptz
)
SELECT
    o.id::text AS key, o.name AS label,
    COALESCE(SUM(f.sold_qty), 0)::bigint AS sold_qty, COALESCE(SUM(f.gross_amount), 0)::bigint AS gross_amount,
    COALESCE(SUM(f.store_share_amount), 0)::bigint AS store_share_amount,
    COALESCE(SUM(f.consignor_share_amount), 0)::bigint AS consignor_share_amount,
    COALESCE(SUM(f.returned_qty), 0)::bigint AS returned_qty, COALESCE(SUM(f.received_qty), 0)::bigint AS received_qty
FROM facts AS f
JOIN scoped AS s ON s.id = f.partnership_id
JOIN organizations AS o
    ON o.id = CASE WHEN s.store_org_id = @organization_id::uuid THEN s.consignor_org_id ELSE s.store_org_id END
GROUP BY o.id, o.name
ORDER BY gross_amount DESC, o.name;

-- name: SalesByProduct :many
WITH scoped AS (
    SELECT p.id, p.consignor_org_id, p.store_org_id
    FROM partnerships AS p
    WHERE p.store_org_id = @organization_id::uuid OR p.consignor_org_id = @organization_id::uuid
), facts AS (
    SELECT
        a.partnership_id, a.product_id, a.occurred_at, a.qty::bigint AS sold_qty, a.gross_amount,
        a.store_share_amount, a.consignor_share_amount, 0::bigint AS returned_qty, 0::bigint AS received_qty
    FROM sale_allocations AS a
    WHERE a.partnership_id IN (SELECT s.id FROM scoped AS s)
        AND a.occurred_at >= @from_at::timestamptz AND a.occurred_at < @to_at::timestamptz
    UNION ALL
    SELECT
        m.partnership_id, m.product_id, m.occurred_at, 0::bigint, 0::bigint, 0::bigint, 0::bigint,
        (CASE WHEN m.type = 'RETURN' THEN -m.qty ELSE 0 END)::bigint,
        (CASE WHEN m.source_type = 'DROP' THEN m.qty ELSE 0 END)::bigint
    FROM stock_movements AS m
    WHERE m.partnership_id IN (SELECT s.id FROM scoped AS s) AND m.type IN ('RETURN', 'RECEIVE', 'ADJUSTMENT')
        AND m.occurred_at >= @from_at::timestamptz AND m.occurred_at < @to_at::timestamptz
)
SELECT
    pr.id::text AS key, (pr.name || ' (' || pr.sku || ')')::text AS label,
    COALESCE(SUM(f.sold_qty), 0)::bigint AS sold_qty, COALESCE(SUM(f.gross_amount), 0)::bigint AS gross_amount,
    COALESCE(SUM(f.store_share_amount), 0)::bigint AS store_share_amount,
    COALESCE(SUM(f.consignor_share_amount), 0)::bigint AS consignor_share_amount,
    COALESCE(SUM(f.returned_qty), 0)::bigint AS returned_qty, COALESCE(SUM(f.received_qty), 0)::bigint AS received_qty
FROM facts AS f
JOIN products AS pr ON pr.id = f.product_id
GROUP BY pr.id, pr.name, pr.sku
ORDER BY gross_amount DESC, pr.name;

-- name: SalesByDay :many
WITH scoped AS (
    SELECT p.id, p.consignor_org_id, p.store_org_id
    FROM partnerships AS p
    WHERE p.store_org_id = @organization_id::uuid OR p.consignor_org_id = @organization_id::uuid
), facts AS (
    SELECT
        a.partnership_id, a.product_id, a.occurred_at, a.qty::bigint AS sold_qty, a.gross_amount,
        a.store_share_amount, a.consignor_share_amount, 0::bigint AS returned_qty, 0::bigint AS received_qty
    FROM sale_allocations AS a
    WHERE a.partnership_id IN (SELECT s.id FROM scoped AS s)
        AND a.occurred_at >= @from_at::timestamptz AND a.occurred_at < @to_at::timestamptz
    UNION ALL
    SELECT
        m.partnership_id, m.product_id, m.occurred_at, 0::bigint, 0::bigint, 0::bigint, 0::bigint,
        (CASE WHEN m.type = 'RETURN' THEN -m.qty ELSE 0 END)::bigint,
        (CASE WHEN m.source_type = 'DROP' THEN m.qty ELSE 0 END)::bigint
    FROM stock_movements AS m
    WHERE m.partnership_id IN (SELECT s.id FROM scoped AS s) AND m.type IN ('RETURN', 'RECEIVE', 'ADJUSTMENT')
        AND m.occurred_at >= @from_at::timestamptz AND m.occurred_at < @to_at::timestamptz
)
SELECT
    to_char(f.occurred_at AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM-DD')::text AS key,
    to_char(f.occurred_at AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM-DD')::text AS label,
    COALESCE(SUM(f.sold_qty), 0)::bigint AS sold_qty, COALESCE(SUM(f.gross_amount), 0)::bigint AS gross_amount,
    COALESCE(SUM(f.store_share_amount), 0)::bigint AS store_share_amount,
    COALESCE(SUM(f.consignor_share_amount), 0)::bigint AS consignor_share_amount,
    COALESCE(SUM(f.returned_qty), 0)::bigint AS returned_qty, COALESCE(SUM(f.received_qty), 0)::bigint AS received_qty
FROM facts AS f
GROUP BY 1, 2
ORDER BY 1;
