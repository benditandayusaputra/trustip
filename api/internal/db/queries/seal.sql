-- name: LockSealHead :one
SELECT seal_seq, COALESCE(seal_head_hash, ''::bytea)::bytea AS seal_head_hash
FROM partnerships
WHERE id = @id::uuid
FOR UPDATE;

-- name: AppendSeal :exec
INSERT INTO seals (
    id, partnership_id, seq, format_version, event_type, entity_type, entity_id, entity_number, content_hash,
    actor_user_id, actor_org_id, occurred_at, prev_hash, entry_hash, signature, key_id, code, created_at
) VALUES (
    @id::uuid, @partnership_id::uuid, @seq::bigint, @format_version::bigint, @event_type::text,
    @entity_type::text, @entity_id::uuid, @entity_number::text, @content_hash::bytea, @actor_user_id::uuid,
    @actor_org_id::uuid, @occurred_at::timestamptz, @prev_hash::bytea, @entry_hash::bytea, @signature::bytea,
    @key_id::text, @code::text, @occurred_at::timestamptz
);

-- name: AdvanceSealHead :exec
UPDATE partnerships
SET seal_seq = @seal_seq::bigint, seal_head_hash = @seal_head_hash::bytea
WHERE id = @id::uuid;

-- name: ListPartnershipSeals :many
SELECT
    id, partnership_id, seq, event_type, entity_type, entity_id, entity_number, content_hash, actor_user_id,
    actor_org_id, occurred_at, prev_hash, entry_hash, signature, key_id, code
FROM seals
WHERE partnership_id = @partnership_id::uuid
ORDER BY seq;

-- name: DeleteSealsOfPartnership :execrows
DELETE FROM seals WHERE partnership_id = @partnership_id::uuid;
