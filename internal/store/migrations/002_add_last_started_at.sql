-- +goose Up
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS last_started_at BIGINT;

-- +goose Down
ALTER TABLE accounts DROP COLUMN IF EXISTS last_started_at;
