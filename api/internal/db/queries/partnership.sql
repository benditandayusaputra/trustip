-- name: CreatePartnership :exec
INSERT INTO partnerships (id, consignor_org_id, store_org_id, status, started_at, created_at, updated_at)
VALUES (
    @id::uuid, @consignor_org_id::uuid, @store_org_id::uuid, 'ACTIVE', @started_at::timestamptz,
    @started_at::timestamptz, @started_at::timestamptz
);
