-- name: ResetDemoData :exec
SELECT reset_demo_data();

-- name: ListUnpaidSettlements :many
SELECT s.id, s.consignor_share_amount, s.issued_at
FROM settlements AS s
WHERE s.partnership_id = @partnership_id::uuid AND s.status = 'ISSUED'
    AND NOT EXISTS (
        SELECT 1 FROM payments AS p WHERE p.settlement_id = s.id AND p.status = 'PENDING_CONFIRMATION'
    )
ORDER BY s.issued_at, s.id;

-- name: ListPendingPayments :many
SELECT p.id, s.issued_at
FROM payments AS p
JOIN settlements AS s ON s.id = p.settlement_id
WHERE p.partnership_id = @partnership_id::uuid AND p.status = 'PENDING_CONFIRMATION'
ORDER BY s.issued_at, p.id;

-- name: GetDemoShape :one
SELECT
    (SELECT count(*) FROM partnerships WHERE status = 'ACTIVE')::bigint AS active_partnerships,
    (SELECT count(*) FROM partnerships WHERE status = 'ENDED')::bigint AS ended_partnerships,
    (SELECT count(*) FROM needs WHERE status = 'OPEN')::bigint AS open_needs,
    (SELECT count(*) FROM needs WHERE status = 'FULFILLED')::bigint AS fulfilled_needs,
    (SELECT count(*) FROM proposals WHERE status = 'OPEN')::bigint AS open_proposals,
    (SELECT count(*) FROM proposals WHERE status = 'DECLINED')::bigint AS declined_proposals,
    (SELECT count(*) FROM settlements WHERE status = 'PAID')::bigint AS paid_settlements,
    (SELECT count(*) FROM payments WHERE status = 'PENDING_CONFIRMATION')::bigint AS pending_payments,
    (SELECT count(*) FROM settlements WHERE status = 'ISSUED')::bigint AS unpaid_settlements,
    (SELECT count(*) FROM settlements WHERE status = 'ISSUED' AND due_date < (@today::text)::date)::bigint
        AS overdue_settlements,
    (SELECT count(*) FROM organizations WHERE is_demo)::bigint AS demo_organizations;
