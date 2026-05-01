-- name: SendOutboxMessage :exec
INSERT INTO outbox (idempotency_key, data, status, kind)
VALUES (
           sqlc.arg(idempotency_key),
           sqlc.arg(data),
           'CREATED'::outbox_status,
           sqlc.arg(kind)
       )
ON CONFLICT (idempotency_key) DO NOTHING;

-- name: GetOutboxMessages :many
UPDATE outbox
SET status = 'IN_PROGRESS'::outbox_status
WHERE idempotency_key IN (
    SELECT idempotency_key
    FROM outbox
    WHERE
        status = 'CREATED'::outbox_status
       OR (
        status = 'IN_PROGRESS'::outbox_status
            AND updated_at < now() - sqlc.arg(in_progress_ttl)::interval
        )
    ORDER BY created_at
    LIMIT sqlc.arg(batch_size)
        FOR UPDATE SKIP LOCKED
)
RETURNING idempotency_key, data, kind;

-- name: MarkOutboxMessagesAsProcessed :exec
UPDATE outbox
SET status = 'SUCCESS'::outbox_status
WHERE idempotency_key = ANY(sqlc.arg(idempotency_keys)::text[]);

-- name: MarkOutboxMessagesAsRetryable :exec
UPDATE outbox
SET status = 'CREATED'::outbox_status
WHERE idempotency_key = ANY(sqlc.arg(idempotency_keys)::text[]);