-- +goose Up
-- Convert updated_at from Unix seconds to Unix milliseconds.
-- The condition guards against rows already stored in milliseconds
-- (values < 10^12 are seconds; after 2001-09-09 all second-timestamps are < 2*10^12).
UPDATE accounts SET updated_at = updated_at * 1000 WHERE updated_at < 1000000000000;

-- +goose Down
UPDATE accounts SET updated_at = updated_at / 1000 WHERE updated_at >= 1000000000000;
