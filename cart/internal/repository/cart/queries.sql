-- name: ListCartByUserId :many
SELECT sku,count
FROM cart.items
WHERE user_id = sqlc.arg(user_id)
ORDER BY sku;

-- name: InsertItem :exec
INSERT INTO cart.items (user_id, sku, count)
VALUES (sqlc.arg(user_id), sqlc.arg(sku), sqlc.arg(count))
ON CONFLICT (user_id, sku) DO UPDATE 
 SET count = cart.items.count + EXCLUDED.count;

-- name: DeleteItemBySku :exec
DELETE FROM cart.items
WHERE sku = sqlc.arg(sku) AND user_id = sqlc.arg(user_id);

-- name: GetCountBySku :one
SELECT count
FROM cart.items
WHERE user_id = sqlc.arg(user_id) AND sku = sqlc.arg(sku);

-- name: ClearCart :exec
DELETE FROM cart.items
WHERE user_id = sqlc.arg(user_id);