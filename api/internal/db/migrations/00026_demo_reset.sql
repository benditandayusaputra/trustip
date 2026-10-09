-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION reset_demo_data() RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public, pg_temp
AS $$
DECLARE
    demo_orgs uuid[] := ARRAY(SELECT id FROM organizations WHERE is_demo);
    demo_users uuid[] := ARRAY(
        SELECT DISTINCT user_id FROM memberships WHERE organization_id = ANY (demo_orgs)
    );
    demo_partnerships uuid[] := ARRAY(
        SELECT id FROM partnerships
        WHERE consignor_org_id = ANY (demo_orgs) AND store_org_id = ANY (demo_orgs)
    );
    demo_proposals uuid[] := ARRAY(
        SELECT id FROM proposals
        WHERE from_org_id = ANY (demo_orgs) OR partnership_id = ANY (demo_partnerships)
    );
BEGIN
    DELETE FROM sale_allocations WHERE partnership_id = ANY (demo_partnerships);
    DELETE FROM return_allocations WHERE count_line_id IN (
        SELECT l.id FROM count_lines AS l JOIN counts AS c ON c.id = l.count_id
        WHERE c.partnership_id = ANY (demo_partnerships)
    );
    DELETE FROM stock_movements WHERE partnership_id = ANY (demo_partnerships);
    DELETE FROM payments WHERE partnership_id = ANY (demo_partnerships);
    DELETE FROM settlements WHERE partnership_id = ANY (demo_partnerships);
    DELETE FROM count_lines WHERE count_id IN (SELECT id FROM counts WHERE partnership_id = ANY (demo_partnerships));
    DELETE FROM counts WHERE partnership_id = ANY (demo_partnerships);
    DELETE FROM drop_items WHERE drop_id IN (SELECT id FROM drops WHERE partnership_id = ANY (demo_partnerships));
    DELETE FROM drops WHERE partnership_id = ANY (demo_partnerships);
    DELETE FROM seals WHERE partnership_id = ANY (demo_partnerships);
    DELETE FROM agreement_items WHERE agreement_id IN (
        SELECT id FROM agreements WHERE partnership_id = ANY (demo_partnerships)
    );
    DELETE FROM agreements WHERE partnership_id = ANY (demo_partnerships);
    DELETE FROM proposal_items WHERE revision_id IN (
        SELECT id FROM proposal_revisions WHERE proposal_id = ANY (demo_proposals)
    );
    DELETE FROM proposal_revisions WHERE proposal_id = ANY (demo_proposals);
    DELETE FROM proposals WHERE id = ANY (demo_proposals);
    DELETE FROM partnerships WHERE id = ANY (demo_partnerships);
    DELETE FROM needs WHERE store_org_id = ANY (demo_orgs);
    DELETE FROM products WHERE consignor_org_id = ANY (demo_orgs);
    DELETE FROM listing_reports WHERE reporter_org_id = ANY (demo_orgs);
    DELETE FROM notifications WHERE user_id = ANY (demo_users);
    DELETE FROM audit_logs WHERE organization_id = ANY (demo_orgs) OR actor_id = ANY (demo_users);
    DELETE FROM document_sequences WHERE organization_id = ANY (demo_orgs);
    DELETE FROM proposal_quotas WHERE organization_id = ANY (demo_orgs);
    DELETE FROM org_stats WHERE organization_id = ANY (demo_orgs);
    DELETE FROM store_accepted_categories WHERE organization_id = ANY (demo_orgs);
    DELETE FROM store_profiles WHERE organization_id = ANY (demo_orgs);
    DELETE FROM consignor_profiles WHERE organization_id = ANY (demo_orgs);
    DELETE FROM org_photos WHERE organization_id = ANY (demo_orgs);
    UPDATE organizations SET logo_file_id = NULL, payment_qr_file_id = NULL WHERE id = ANY (demo_orgs);
    DELETE FROM files WHERE organization_id = ANY (demo_orgs);
    DELETE FROM memberships WHERE organization_id = ANY (demo_orgs);
    DELETE FROM organizations WHERE id = ANY (demo_orgs);
    DELETE FROM users WHERE id = ANY (demo_users);
END;
$$;
-- +goose StatementEnd

-- +goose Down
DROP FUNCTION reset_demo_data();
