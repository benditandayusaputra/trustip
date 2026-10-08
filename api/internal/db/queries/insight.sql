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
