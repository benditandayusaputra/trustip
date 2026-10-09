-- name: GetPublicSeal :one
SELECT
    s.id, s.partnership_id, s.seq, s.event_type, s.entity_type, s.entity_id, s.entity_number, s.content_hash,
    COALESCE(s.actor_user_id::text, '')::text AS actor_user_id,
    COALESCE(s.actor_org_id::text, '')::text AS actor_org_id,
    s.occurred_at, s.prev_hash, s.entry_hash, s.signature, s.key_id, s.code,
    consignor.name AS consignor_name,
    store.name AS store_name,
    COALESCE(previous.entry_hash, ''::bytea)::bytea AS previous_entry_hash
FROM seals AS s
JOIN partnerships AS p ON p.id = s.partnership_id
JOIN organizations AS consignor ON consignor.id = p.consignor_org_id
JOIN organizations AS store ON store.id = p.store_org_id
LEFT JOIN seals AS previous ON previous.partnership_id = s.partnership_id AND previous.seq = s.seq - 1
WHERE s.code = @code::text;

-- name: ListSealedPartnerships :many
SELECT p.id
FROM partnerships AS p
WHERE p.seal_seq > 0
ORDER BY p.id;
