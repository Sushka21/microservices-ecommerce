-- name: SendOutboxMessage :exec
INSERT INTO inbox (idempotency_key, data, status, kind)
VALUES (
           sqlc.arg(idempotency_key),
           sqlc.arg(data),
           'CREATED'::inbox_status,
           sqlc.arg(kind)
       )
ON CONFLICT (idempotency_key) DO NOTHING;

-- name: GetOutboxMessages :many
UPDATE inbox
SET status = 'IN_PROGRESS'::inbox_status
WHERE idempotency_key IN (
    SELECT idempotency_key
    FROM inbox
    WHERE
        status = 'CREATED'::inbox_status
       OR (
        status = 'IN_PROGRESS'::inbox_status
            AND updated_at < now() - sqlc.arg(in_progress_ttl)::interval
        )
    ORDER BY created_at
    LIMIT sqlc.arg(batch_size)
        FOR UPDATE SKIP LOCKED
)
RETURNING idempotency_key, data, kind;

-- name: MarkOutboxMessagesAsProcessed :exec
UPDATE inbox
SET status = 'SUCCESS'::inbox_status
WHERE idempotency_key = ANY(sqlc.arg(idempotency_keys)::text[]);

-- name: MarkOutboxMessagesAsRetryable :exec
UPDATE inbox
SET status = 'CREATED'::inbox_status
WHERE idempotency_key = ANY(sqlc.arg(idempotency_keys)::text[]);