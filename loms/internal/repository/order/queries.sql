-- name: InsertOrder :one
INSERT INTO loms.orders (user_id, status)
VALUES (sqlc.arg(user_id), sqlc.arg(status)::loms.order_status)
RETURNING id, user_id, status, created_at, updated_at;

-- name: InsertOrderItem :exec
INSERT INTO loms.order_items (order_id, sku, count)
VALUES (sqlc.arg(order_id), sqlc.arg(sku), sqlc.arg(count));

-- name: GetOrderByID :one
SELECT id, user_id, status, created_at, updated_at
FROM loms.orders
WHERE id = sqlc.arg(id);

-- name: GetOrderForUpdateByID :one
SELECT id, user_id, status, created_at, updated_at
FROM loms.orders
WHERE id = sqlc.arg(id)
FOR UPDATE;

-- name: ListOrderItemsByOrderID :many
SELECT sku, count
FROM loms.order_items
WHERE order_id = sqlc.arg(id)
ORDER BY sku;

-- name: SetOrderStatus :exec
UPDATE loms.orders
SET status = sqlc.arg(status)::loms.order_status
WHERE id = sqlc.arg(id);    