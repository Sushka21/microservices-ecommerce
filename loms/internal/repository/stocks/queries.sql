-- name: DecrementAvailableStock :execrows
UPDATE loms.available_stocks
SET count = count - sqlc.arg(count)
WHERE sku = sqlc.arg(sku)
  AND count >= sqlc.arg(count);

-- name: AddAvailableStock :exec
INSERT INTO loms.available_stocks (sku, count)
VALUES (sqlc.arg(sku), sqlc.arg(count))
ON CONFLICT (sku) DO UPDATE
  SET count = loms.available_stocks.count + EXCLUDED.count;

-- name: UpsertAvailableStock :exec
INSERT INTO loms.available_stocks (sku, count)
VALUES (sqlc.arg(sku), sqlc.arg(count))
ON CONFLICT (sku) DO UPDATE
SET count = EXCLUDED.count;

-- name: GetAvailableStockBySku :one
SELECT  count
FROM loms.available_stocks
WHERE sku = sqlc.arg(sku);

-- name: UpsertReservedStock :exec
INSERT INTO loms.reserved_stocks (sku, order_id,count)
VALUES (sqlc.arg(sku),sqlc.arg(order_id) ,sqlc.arg(count))
ON CONFLICT (sku,order_id) DO UPDATE
  SET count = loms.reserved_stocks.count + EXCLUDED.count;

-- name: DecrementReservedStock :execrows
WITH updated AS (
    UPDATE loms.reserved_stocks AS rs
    SET count = rs.count - sqlc.arg(count)
    WHERE rs.sku = sqlc.arg(sku)
      AND rs.order_id = sqlc.arg(order_id)
      AND rs.count >= sqlc.arg(count)
    RETURNING rs.sku, rs.order_id, rs.count
)
DELETE FROM loms.reserved_stocks
WHERE (sku, order_id) IN (
    SELECT sku, order_id
    FROM updated
    WHERE count = 0
);
