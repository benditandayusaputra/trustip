-- name: TamperSealNumber :execrows
UPDATE seals SET entity_number = @entity_number::text WHERE id = @id::uuid;

-- name: TamperAgreementItemPrice :execrows
UPDATE agreement_items SET retail_price = @retail_price::bigint
WHERE agreement_id = @agreement_id::uuid AND product_id = @product_id::uuid;
