-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE inbox_status as ENUM ('CREATED', 'IN_PROGRESS', 'SUCCESS');

CREATE TABLE inbox
(
    idempotency_key TEXT PRIMARY KEY,
    data            JSONB                   NOT NULL,
    status          inbox_status           NOT NULL,
    kind            INT                     NOT NULL,
    created_at      TIMESTAMP DEFAULT now() NOT NULL,
    updated_at      TIMESTAMP DEFAULT now() NOT NULL
);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_inbox_timestamp() RETURNS TRIGGER AS
$$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd


CREATE OR REPLACE TRIGGER trigger_update_inbox_timestamp
    BEFORE UPDATE
    ON inbox
    FOR EACH ROW
EXECUTE FUNCTION update_inbox_timestamp();


-- +goose Down
DROP TRIGGER IF EXISTS trigger_update_inbox_timestamp ON inbox;
DROP FUNCTION IF EXISTS update_inbox_timestamp;
DROP TABLE IF EXISTS inbox;
DROP TYPE IF EXISTS inbox_status;