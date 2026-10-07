-- name: TamperSealNumber :execrows
UPDATE seals SET entity_number = @entity_number::text WHERE id = @id::uuid;
