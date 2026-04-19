-- +goose Up
ALTER TABLE refresh_tokens
ADD COLUMN jti UUID NOT NULL UNIQUE;

-- +goose Down
ALTER TABLE refresh_tokens
DROP COLUMN jti;