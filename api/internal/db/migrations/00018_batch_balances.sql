-- +goose Up
CREATE VIEW batch_balances AS
SELECT
    i.id AS drop_item_id, d.partnership_id, i.product_id, i.batch_code, i.expiry_date,
    d.occurred_at AS received_at, COALESCE(sum(m.qty), 0)::integer AS balance
FROM drop_items AS i
JOIN drops AS d ON d.id = i.drop_id
LEFT JOIN stock_movements AS m ON m.drop_item_id = i.id
GROUP BY i.id, d.partnership_id, i.product_id, i.batch_code, i.expiry_date, d.occurred_at;

-- +goose Down
DROP VIEW batch_balances;
