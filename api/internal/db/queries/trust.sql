-- name: ListReputationFacts :many
SELECT
    o.id AS organization_id,
    (
        SELECT count(*) FROM partnerships AS p
        WHERE p.status = 'ACTIVE' AND (p.store_org_id = o.id OR p.consignor_org_id = o.id)
    )::bigint AS active_partners,
    (
        SELECT count(*) FROM settlements AS st
        WHERE st.store_org_id = o.id
            AND st.due_date >= CAST(@since::text AS date) AND st.due_date < CAST(@today::text AS date)
    )::bigint AS due_settlements,
    (
        SELECT count(*) FROM settlements AS st
        WHERE st.store_org_id = o.id
            AND st.due_date >= CAST(@since::text AS date) AND st.due_date < CAST(@today::text AS date)
            AND EXISTS (
                SELECT 1 FROM payments AS pm
                WHERE pm.settlement_id = st.id AND pm.status = 'CONFIRMED' AND pm.paid_on <= st.due_date
            )
    )::bigint AS on_time_settlements,
    (
        SELECT count(*) FROM products AS pr WHERE pr.consignor_org_id = o.id AND NOT pr.is_archived
    )::bigint AS products
FROM organizations AS o
ORDER BY o.id;

-- name: SaveOrgStats :execrows
INSERT INTO org_stats (
    organization_id, active_partner_count, on_time_payment_bps, settlements_counted, product_count, computed_at
)
SELECT
    unnest(@organization_ids::uuid[]), unnest(@active_partner_counts::bigint[])::integer,
    unnest(@on_time_payment_bps::bigint[])::integer, unnest(@settlements_counted::bigint[])::integer,
    unnest(@product_counts::bigint[])::integer, @computed_at::timestamptz
ON CONFLICT (organization_id) DO UPDATE SET
    active_partner_count = EXCLUDED.active_partner_count,
    on_time_payment_bps = EXCLUDED.on_time_payment_bps,
    settlements_counted = EXCLUDED.settlements_counted,
    product_count = EXCLUDED.product_count,
    computed_at = EXCLUDED.computed_at;
