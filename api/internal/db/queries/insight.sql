-- name: FindInsightPartnership :one
SELECT p.id
FROM partnerships AS p
WHERE p.id = @partnership_id::uuid
    AND (p.store_org_id = @organization_id::uuid OR p.consignor_org_id = @organization_id::uuid);

-- name: ListDropSuggestionProducts :many
SELECT
    ai.product_id, a.count_rhythm,
    COALESCE(pr.shelf_life_days, 0)::bigint AS shelf_life_days,
    COALESCE((
        SELECT sum(b.balance) FROM batch_balances AS b
        WHERE b.partnership_id = p.id AND b.product_id = ai.product_id
    ), 0)::bigint AS stock_units,
    COALESCE((
        SELECT sum(s.qty) FROM sale_allocations AS s
        WHERE s.partnership_id = p.id AND s.product_id = ai.product_id AND s.occurred_at >= @sales_since::timestamptz
    ), 0)::bigint AS sold_since,
    COALESCE((
        SELECT min(d.occurred_at) FROM drops AS d
        WHERE d.partnership_id = p.id AND d.status <> 'CANCELLED'
    ), @now::timestamptz)::timestamptz AS first_drop_at
FROM partnerships AS p
JOIN agreements AS a ON a.partnership_id = p.id AND a.status = 'ACTIVE'
JOIN agreement_items AS ai ON ai.agreement_id = a.id AND ai.is_active
JOIN products AS pr ON pr.id = ai.product_id
WHERE p.id = @partnership_id::uuid
    AND (p.store_org_id = @organization_id::uuid OR p.consignor_org_id = @organization_id::uuid)
ORDER BY ai.product_id;

-- name: ListSameWeekdaySales :many
WITH daily AS (
    SELECT
        l.product_id, (c.counted_at AT TIME ZONE 'Asia/Jakarta')::date AS sales_date,
        sum(l.sold_qty)::bigint AS sold, bool_or(l.counted_qty = 0) AS sold_out
    FROM counts AS c
    JOIN count_lines AS l ON l.count_id = c.id
    JOIN partnerships AS p ON p.id = c.partnership_id
    WHERE c.partnership_id = @partnership_id::uuid
        AND (p.store_org_id = @organization_id::uuid OR p.consignor_org_id = @organization_id::uuid)
        AND extract(isodow FROM c.counted_at AT TIME ZONE 'Asia/Jakarta') = @iso_weekday::bigint
    GROUP BY l.product_id, sales_date
),
ranked AS (
    SELECT
        product_id, sales_date, sold, sold_out,
        row_number() OVER (PARTITION BY product_id ORDER BY sales_date DESC) AS occurrence
    FROM daily
)
SELECT product_id, sold, sold_out
FROM ranked
WHERE occurrence <= @max_occurrences::bigint
ORDER BY product_id, sales_date DESC;

-- name: ListInsightStock :many
WITH scoped AS (
    SELECT
        p.id, p.status,
        CASE WHEN p.store_org_id = @organization_id::uuid THEN p.consignor_org_id ELSE p.store_org_id END
            AS counterpart_id
    FROM partnerships AS p
    WHERE p.store_org_id = @organization_id::uuid OR p.consignor_org_id = @organization_id::uuid
),
held AS (
    SELECT b.partnership_id, b.product_id
    FROM batch_balances AS b
    WHERE b.partnership_id IN (SELECT s.id FROM scoped AS s) AND b.balance > 0
    UNION
    SELECT a.partnership_id, ai.product_id
    FROM agreements AS a
    JOIN agreement_items AS ai ON ai.agreement_id = a.id AND ai.is_active
    JOIN scoped AS s ON s.id = a.partnership_id AND s.status = 'ACTIVE'
    WHERE a.status = 'ACTIVE'
)
SELECT
    h.partnership_id, o.name AS counterpart_name, COALESCE(a.count_rhythm, 'EACH_VISIT')::text AS count_rhythm,
    h.product_id, pr.name AS product_name, COALESCE(pr.shelf_life_days, 0)::bigint AS shelf_life_days,
    COALESCE(ai.retail_price, (
        SELECT i.retail_price FROM drop_items AS i
        JOIN drops AS d ON d.id = i.drop_id
        WHERE d.partnership_id = h.partnership_id AND i.product_id = h.product_id
        ORDER BY d.occurred_at DESC, i.id DESC
        LIMIT 1
    ), 0)::bigint AS retail_price,
    COALESCE(b.drop_item_id::text, '')::text AS batch_id, COALESCE(b.batch_code, '')::text AS batch_code,
    COALESCE(to_char(b.expiry_date, 'YYYY-MM-DD'), '')::text AS expiry_date, COALESCE(b.balance, 0)::bigint AS balance,
    COALESCE((
        SELECT sum(sa.qty) FROM sale_allocations AS sa
        WHERE sa.partnership_id = h.partnership_id AND sa.product_id = h.product_id
            AND sa.occurred_at >= @sales_since::timestamptz
    ), 0)::bigint AS sold_since,
    COALESCE((
        SELECT min(d.occurred_at) FROM drops AS d
        WHERE d.partnership_id = h.partnership_id AND d.status <> 'CANCELLED'
    ), @now::timestamptz)::timestamptz AS first_drop_at,
    COALESCE((
        SELECT max(c.counted_at) FROM counts AS c WHERE c.partnership_id = h.partnership_id
    ), (
        SELECT min(d.occurred_at) FROM drops AS d
        WHERE d.partnership_id = h.partnership_id AND d.status <> 'CANCELLED'
    ), @now::timestamptz)::timestamptz AS uncounted_since
FROM held AS h
JOIN scoped AS s ON s.id = h.partnership_id
JOIN organizations AS o ON o.id = s.counterpart_id
JOIN products AS pr ON pr.id = h.product_id
LEFT JOIN agreements AS a ON a.partnership_id = h.partnership_id AND a.status = 'ACTIVE'
LEFT JOIN agreement_items AS ai ON ai.agreement_id = a.id AND ai.product_id = h.product_id
LEFT JOIN batch_balances AS b ON b.partnership_id = h.partnership_id AND b.product_id = h.product_id AND b.balance > 0
ORDER BY h.partnership_id, h.product_id, b.expiry_date NULLS LAST, b.drop_item_id;

-- name: ListOpenSettlements :many
SELECT
    st.id, st.number, st.partnership_id, o.name AS counterpart_name,
    to_char(st.due_date, 'YYYY-MM-DD')::text AS due_date, st.consignor_share_amount
FROM settlements AS st
JOIN organizations AS o
    ON o.id = CASE WHEN st.store_org_id = @organization_id::uuid THEN st.consignor_org_id ELSE st.store_org_id END
WHERE (st.store_org_id = @organization_id::uuid OR st.consignor_org_id = @organization_id::uuid)
    AND st.status = 'ISSUED'
ORDER BY st.due_date, st.id;

-- name: GetDashboardFigures :one
WITH scoped AS (
    SELECT p.id, p.status, p.consignor_org_id
    FROM partnerships AS p
    WHERE p.store_org_id = @organization_id::uuid OR p.consignor_org_id = @organization_id::uuid
),
open_settlements AS (
    SELECT st.id, st.due_date, st.consignor_share_amount
    FROM settlements AS st
    WHERE (st.store_org_id = @organization_id::uuid OR st.consignor_org_id = @organization_id::uuid)
        AND st.status = 'ISSUED'
)
SELECT
    (SELECT count(*) FROM scoped AS s WHERE s.status = 'ACTIVE')::bigint AS active_partners,
    COALESCE((
        SELECT sum(b.balance) FROM batch_balances AS b WHERE b.partnership_id IN (SELECT s.id FROM scoped AS s)
    ), 0)::bigint AS stock_units,
    (SELECT count(*) FROM open_settlements)::bigint AS unpaid_count,
    COALESCE((SELECT sum(o.consignor_share_amount) FROM open_settlements AS o), 0)::bigint AS unpaid_amount,
    (SELECT count(*) FROM open_settlements AS o WHERE o.due_date < CAST(@today::text AS date))::bigint
        AS overdue_count,
    COALESCE((
        SELECT sum(o.consignor_share_amount) FROM open_settlements AS o
        WHERE o.due_date < CAST(@today::text AS date)
    ), 0)::bigint AS overdue_amount,
    (
        SELECT count(*) FROM open_settlements AS o
        WHERE o.due_date BETWEEN CAST(@today::text AS date) AND CAST(@due_soon_until::text AS date)
    )::bigint AS due_soon_count,
    COALESCE((
        SELECT sum(o.consignor_share_amount) FROM open_settlements AS o
        WHERE o.due_date BETWEEN CAST(@today::text AS date) AND CAST(@due_soon_until::text AS date)
    ), 0)::bigint AS due_soon_amount,
    (
        SELECT count(*) FROM drops AS d
        JOIN scoped AS s ON s.id = d.partnership_id AND s.consignor_org_id = @organization_id::uuid
        WHERE d.status = 'RECORDED' AND d.received_at > @pending_since::timestamptz
    )::bigint AS drops_to_confirm,
    (
        SELECT count(*) FROM counts AS c
        JOIN scoped AS s ON s.id = c.partnership_id AND s.consignor_org_id = @organization_id::uuid
        WHERE c.status = 'RECORDED' AND c.received_at > @pending_since::timestamptz
    )::bigint AS counts_to_confirm,
    (
        SELECT count(*) FROM payments AS pm
        JOIN scoped AS s ON s.id = pm.partnership_id AND s.consignor_org_id = @organization_id::uuid
        WHERE pm.status = 'PENDING_CONFIRMATION'
    )::bigint AS payments_to_review,
    (
        SELECT count(*) FROM open_settlements AS o
        WHERE EXISTS (SELECT 1 FROM payments AS pm WHERE pm.settlement_id = o.id AND pm.status = 'REJECTED')
            AND NOT EXISTS (
                SELECT 1 FROM payments AS pm WHERE pm.settlement_id = o.id AND pm.status <> 'REJECTED'
            )
    )::bigint AS payments_rejected,
    (
        SELECT count(*) FROM proposals AS pp
        WHERE pp.status = 'OPEN' AND pp.awaiting_org_id = @organization_id::uuid
    )::bigint AS proposals_awaiting;

-- name: ListUncountedPartners :many
SELECT p.id AS partnership_id, o.name AS counterpart_name, stock.units AS stock_units
FROM partnerships AS p
JOIN organizations AS o ON o.id = p.consignor_org_id
JOIN LATERAL (
    SELECT COALESCE(sum(b.balance), 0)::bigint AS units FROM batch_balances AS b WHERE b.partnership_id = p.id
) AS stock ON true
WHERE p.store_org_id = @organization_id::uuid AND p.status = 'ACTIVE' AND stock.units > 0
    AND NOT EXISTS (
        SELECT 1 FROM counts AS c WHERE c.partnership_id = p.id AND c.counted_at >= @today_start::timestamptz
    )
ORDER BY o.name, p.id;
