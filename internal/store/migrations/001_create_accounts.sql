-- +goose Up
CREATE TABLE IF NOT EXISTS accounts (
    username        TEXT PRIMARY KEY,
    config_json     TEXT NOT NULL,
    enabled         BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at      BIGINT NOT NULL
);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION accounts_notify() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN PERFORM pg_notify('accounts_changed', ''); RETURN NULL; END;
$$;
-- +goose StatementEnd

DROP TRIGGER IF EXISTS accounts_changed_trigger ON accounts;
CREATE TRIGGER accounts_changed_trigger
    AFTER INSERT OR UPDATE OR DELETE ON accounts
    FOR EACH STATEMENT EXECUTE FUNCTION accounts_notify();

-- +goose Down
DROP TRIGGER IF EXISTS accounts_changed_trigger ON accounts;
DROP FUNCTION IF EXISTS accounts_notify();
DROP TABLE IF EXISTS accounts;
