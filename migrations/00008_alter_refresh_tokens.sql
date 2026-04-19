-- +goose Up
ALTER TABLE refresh_tokens
DROP COLUMN revoked;

-- +goose Down
ALTER TABLE refresh_tokens
ADD COLUMN revoked BOOLEAN NOT NULL DEFAULT FALSE;
