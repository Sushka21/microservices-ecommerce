-- name: GetProductBySKU :one
SELECT sku, names, price
FROM loms.products
WHERE sku = sqlc.arg(sku);

-- name: CreateProduct :one
INSERT INTO loms.products (names,price)
VALUES (sqlc.arg(names), sqlc.arg(price))
RETURNING sku, names, price;

-- name: ListProductBySkus :many
SELECT sku,names, price
FROM loms.products
WHERE sku = ANY(sqlc.arg(skus)::integer[]);