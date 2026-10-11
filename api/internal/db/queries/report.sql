-- name: SalesByPartner :many
WITH scoped AS (
    SELECT p.id, p.consignor_org_id, p.store_org_id
    FROM partnerships AS p
    WHERE (p.store_org_id = @organization_id::uuid OR p.consignor_org_id = @organization_id::uuid)
        AND (
            @partner_id::text = ''
            OR (CASE WHEN p.store_org_id = @organization_id::uuid THEN p.consignor_org_id ELSE p.store_org_id END)::text
                = @partner_id::text
        )
), facts AS (
    SELECT
        a.partnership_id, a.product_id, a.occurred_at, a.qty::bigint AS sold_qty, a.gross_amount,
        a.store_share_amount, a.consignor_share_amount, 0::bigint AS returned_qty, 0::bigint AS received_qty
    FROM sale_allocations AS a
    WHERE a.partnership_id IN (SELECT s.id FROM scoped AS s)
        AND a.occurred_at >= @from_at::timestamptz AND a.occurred_at < @to_at::timestamptz
        AND (@product_id::text = '' OR a.product_id::text = @product_id::text)
    UNION ALL
    SELECT
        m.partnership_id, m.product_id, m.occurred_at, 0::bigint, 0::bigint, 0::bigint, 0::bigint,
        (CASE WHEN m.type = 'RETURN' THEN -m.qty ELSE 0 END)::bigint,
        (CASE WHEN m.source_type = 'DROP' THEN m.qty ELSE 0 END)::bigint
    FROM stock_movements AS m
    WHERE m.partnership_id IN (SELECT s.id FROM scoped AS s) AND m.type IN ('RETURN', 'RECEIVE', 'ADJUSTMENT')
        AND m.occurred_at >= @from_at::timestamptz AND m.occurred_at < @to_at::timestamptz
        AND (@product_id::text = '' OR m.product_id::text = @product_id::text)
)
SELECT
    o.id::text AS key, o.name AS label,
    ''::text AS photo_key, ''::text AS consignor_slug, ''::text AS sku,
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
    WHERE (p.store_org_id = @organization_id::uuid OR p.consignor_org_id = @organization_id::uuid)
        AND (
            @partner_id::text = ''
            OR (CASE WHEN p.store_org_id = @organization_id::uuid THEN p.consignor_org_id ELSE p.store_org_id END)::text
                = @partner_id::text
        )
), facts AS (
    SELECT
        a.partnership_id, a.product_id, a.occurred_at, a.qty::bigint AS sold_qty, a.gross_amount,
        a.store_share_amount, a.consignor_share_amount, 0::bigint AS returned_qty, 0::bigint AS received_qty
    FROM sale_allocations AS a
    WHERE a.partnership_id IN (SELECT s.id FROM scoped AS s)
        AND a.occurred_at >= @from_at::timestamptz AND a.occurred_at < @to_at::timestamptz
        AND (@product_id::text = '' OR a.product_id::text = @product_id::text)
    UNION ALL
    SELECT
        m.partnership_id, m.product_id, m.occurred_at, 0::bigint, 0::bigint, 0::bigint, 0::bigint,
        (CASE WHEN m.type = 'RETURN' THEN -m.qty ELSE 0 END)::bigint,
        (CASE WHEN m.source_type = 'DROP' THEN m.qty ELSE 0 END)::bigint
    FROM stock_movements AS m
    WHERE m.partnership_id IN (SELECT s.id FROM scoped AS s) AND m.type IN ('RETURN', 'RECEIVE', 'ADJUSTMENT')
        AND m.occurred_at >= @from_at::timestamptz AND m.occurred_at < @to_at::timestamptz
        AND (@product_id::text = '' OR m.product_id::text = @product_id::text)
)
SELECT
    pr.id::text AS key, (pr.name || ' (' || pr.sku || ')')::text AS label,
    COALESCE(image.storage_key, '')::text AS photo_key, owner.slug::text AS consignor_slug, pr.sku::text AS sku,
    COALESCE(SUM(f.sold_qty), 0)::bigint AS sold_qty, COALESCE(SUM(f.gross_amount), 0)::bigint AS gross_amount,
    COALESCE(SUM(f.store_share_amount), 0)::bigint AS store_share_amount,
    COALESCE(SUM(f.consignor_share_amount), 0)::bigint AS consignor_share_amount,
    COALESCE(SUM(f.returned_qty), 0)::bigint AS returned_qty, COALESCE(SUM(f.received_qty), 0)::bigint AS received_qty
FROM facts AS f
JOIN products AS pr ON pr.id = f.product_id
JOIN organizations AS owner ON owner.id = pr.consignor_org_id
LEFT JOIN files AS image ON image.id = pr.photo_file_id
GROUP BY pr.id, pr.name, pr.sku, owner.slug, image.storage_key
ORDER BY gross_amount DESC, pr.name;

-- name: SalesByDay :many
WITH scoped AS (
    SELECT p.id, p.consignor_org_id, p.store_org_id
    FROM partnerships AS p
    WHERE (p.store_org_id = @organization_id::uuid OR p.consignor_org_id = @organization_id::uuid)
        AND (
            @partner_id::text = ''
            OR (CASE WHEN p.store_org_id = @organization_id::uuid THEN p.consignor_org_id ELSE p.store_org_id END)::text
                = @partner_id::text
        )
), facts AS (
    SELECT
        a.partnership_id, a.product_id, a.occurred_at, a.qty::bigint AS sold_qty, a.gross_amount,
        a.store_share_amount, a.consignor_share_amount, 0::bigint AS returned_qty, 0::bigint AS received_qty
    FROM sale_allocations AS a
    WHERE a.partnership_id IN (SELECT s.id FROM scoped AS s)
        AND a.occurred_at >= @from_at::timestamptz AND a.occurred_at < @to_at::timestamptz
        AND (@product_id::text = '' OR a.product_id::text = @product_id::text)
    UNION ALL
    SELECT
        m.partnership_id, m.product_id, m.occurred_at, 0::bigint, 0::bigint, 0::bigint, 0::bigint,
        (CASE WHEN m.type = 'RETURN' THEN -m.qty ELSE 0 END)::bigint,
        (CASE WHEN m.source_type = 'DROP' THEN m.qty ELSE 0 END)::bigint
    FROM stock_movements AS m
    WHERE m.partnership_id IN (SELECT s.id FROM scoped AS s) AND m.type IN ('RETURN', 'RECEIVE', 'ADJUSTMENT')
        AND m.occurred_at >= @from_at::timestamptz AND m.occurred_at < @to_at::timestamptz
        AND (@product_id::text = '' OR m.product_id::text = @product_id::text)
)
SELECT
    to_char(f.occurred_at AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM-DD')::text AS key,
    to_char(f.occurred_at AT TIME ZONE 'Asia/Jakarta', 'YYYY-MM-DD')::text AS label,
    ''::text AS photo_key, ''::text AS consignor_slug, ''::text AS sku,
    COALESCE(SUM(f.sold_qty), 0)::bigint AS sold_qty, COALESCE(SUM(f.gross_amount), 0)::bigint AS gross_amount,
    COALESCE(SUM(f.store_share_amount), 0)::bigint AS store_share_amount,
    COALESCE(SUM(f.consignor_share_amount), 0)::bigint AS consignor_share_amount,
    COALESCE(SUM(f.returned_qty), 0)::bigint AS returned_qty, COALESCE(SUM(f.received_qty), 0)::bigint AS received_qty
FROM facts AS f
GROUP BY 1, 2
ORDER BY 1;

-- name: DailyRecap :many
WITH scoped AS (
    SELECT p.id, p.consignor_org_id
    FROM partnerships AS p
    WHERE p.store_org_id = @store_org_id::uuid
), facts AS (
    SELECT
        a.partnership_id, a.qty::bigint AS sold_qty, a.gross_amount, a.store_share_amount,
        a.consignor_share_amount, 0::bigint AS returned_qty, 0::bigint AS received_qty,
        (CASE WHEN st.status = 'PAID' THEN a.consignor_share_amount ELSE 0 END)::bigint AS paid_amount
    FROM sale_allocations AS a
    LEFT JOIN settlements AS st ON st.id = a.settlement_id
    WHERE a.partnership_id IN (SELECT s.id FROM scoped AS s)
        AND a.occurred_at >= @from_at::timestamptz AND a.occurred_at < @to_at::timestamptz
    UNION ALL
    SELECT
        m.partnership_id, 0::bigint, 0::bigint, 0::bigint, 0::bigint,
        (CASE WHEN m.type = 'RETURN' THEN -m.qty ELSE 0 END)::bigint,
        (CASE WHEN m.source_type = 'DROP' THEN m.qty ELSE 0 END)::bigint, 0::bigint
    FROM stock_movements AS m
    WHERE m.partnership_id IN (SELECT s.id FROM scoped AS s) AND m.type IN ('RETURN', 'RECEIVE', 'ADJUSTMENT')
        AND m.occurred_at >= @from_at::timestamptz AND m.occurred_at < @to_at::timestamptz
)
SELECT
    o.id::text AS partner_id, o.name,
    COALESCE(SUM(f.sold_qty), 0)::bigint AS sold_qty, COALESCE(SUM(f.gross_amount), 0)::bigint AS gross_amount,
    COALESCE(SUM(f.store_share_amount), 0)::bigint AS store_share_amount,
    COALESCE(SUM(f.consignor_share_amount), 0)::bigint AS consignor_share_amount,
    COALESCE(SUM(f.returned_qty), 0)::bigint AS returned_qty, COALESCE(SUM(f.received_qty), 0)::bigint AS received_qty,
    COALESCE(SUM(f.paid_amount), 0)::bigint AS paid_amount
FROM facts AS f
JOIN scoped AS s ON s.id = f.partnership_id
JOIN organizations AS o ON o.id = s.consignor_org_id
GROUP BY o.id, o.name
ORDER BY o.name, o.id;

-- name: SettlementReport :many
SELECT
    o.id::text AS partner_id, o.name,
    COUNT(*)::bigint AS issued_count, COALESCE(SUM(st.consignor_share_amount), 0)::bigint AS issued_amount,
    COUNT(*) FILTER (WHERE st.status = 'PAID')::bigint AS paid_count,
    COALESCE(SUM(st.consignor_share_amount) FILTER (WHERE st.status = 'PAID'), 0)::bigint AS paid_amount,
    COUNT(*) FILTER (WHERE st.status = 'ISSUED')::bigint AS unpaid_count,
    COALESCE(SUM(st.consignor_share_amount) FILTER (WHERE st.status = 'ISSUED'), 0)::bigint AS unpaid_amount,
    COUNT(*) FILTER (WHERE st.status = 'ISSUED' AND st.due_date < CAST(@today::text AS date))::bigint
        AS overdue_count,
    COALESCE(SUM(st.consignor_share_amount) FILTER (
        WHERE st.status = 'ISSUED' AND st.due_date < CAST(@today::text AS date)
    ), 0)::bigint AS overdue_amount
FROM settlements AS st
JOIN organizations AS o
    ON o.id = CASE WHEN st.store_org_id = @organization_id::uuid THEN st.consignor_org_id ELSE st.store_org_id END
WHERE (st.store_org_id = @organization_id::uuid OR st.consignor_org_id = @organization_id::uuid)
    AND st.issued_at >= @from_at::timestamptz AND st.issued_at < @to_at::timestamptz
GROUP BY o.id, o.name
ORDER BY o.name, o.id;
