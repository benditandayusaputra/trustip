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

-- name: SaveListingReport :one
INSERT INTO listing_reports (
    id, reporter_user_id, reporter_org_id, target_type, target_id, reason, note, created_at, updated_at
)
SELECT
    @id::uuid, @reporter_user_id::uuid, @reporter_org_id::uuid, @target_type::text, @target_id::uuid,
    @reason::text, @note::text, @reported_at::timestamptz, @reported_at::timestamptz
WHERE CASE @target_type::text
    WHEN 'ORGANIZATION' THEN EXISTS (
        SELECT 1 FROM organizations AS o WHERE o.id = @target_id::uuid AND o.id <> @reporter_org_id::uuid
    )
    WHEN 'PRODUCT' THEN EXISTS (
        SELECT 1 FROM products AS p
        WHERE p.id = @target_id::uuid AND p.consignor_org_id <> @reporter_org_id::uuid AND NOT p.is_archived
    )
    WHEN 'NEED' THEN EXISTS (
        SELECT 1 FROM needs AS n WHERE n.id = @target_id::uuid AND n.store_org_id <> @reporter_org_id::uuid
    )
    ELSE false
END
ON CONFLICT (reporter_user_id, target_type, target_id) WHERE status = 'OPEN' DO UPDATE SET
    reason = EXCLUDED.reason, note = EXCLUDED.note, updated_at = EXCLUDED.updated_at
RETURNING id, status, created_at;

-- name: SetOrganizationStatus :execrows
UPDATE organizations SET status = @status::text, updated_at = @updated_at::timestamptz
WHERE id = @id::uuid;
