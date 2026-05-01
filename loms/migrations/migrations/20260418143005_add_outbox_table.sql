-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE outbox_status as ENUM ('CREATED', 'IN_PROGRESS', 'SUCCESS');

CREATE TABLE outbox
(
    idempotency_key TEXT PRIMARY KEY,
    data            JSONB                   NOT NULL,
    status          outbox_status           NOT NULL,
    kind            INT                     NOT NULL,
    created_at      TIMESTAMP DEFAULT now() NOT NULL,
    updated_at      TIMESTAMP DEFAULT now() NOT NULL
);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_outbox_timestamp() RETURNS TRIGGER AS
$$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd


CREATE OR REPLACE TRIGGER trigger_update_outbox_timestamp
    BEFORE UPDATE
    ON outbox
    FOR EACH ROW
EXECUTE FUNCTION update_outbox_timestamp();


-- +goose Down
DROP TRIGGER IF EXISTS trigger_update_outbox_timestamp ON outbox;
DROP FUNCTION IF EXISTS update_outbox_timestamp;
DROP TABLE IF EXISTS outbox;
DROP TYPE IF EXISTS outbox_status;