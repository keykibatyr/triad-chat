-- +goose Up
CREATE TYPE sender AS ENUM 
('user', 'ai', 'system');

ALTER TABLE messages
ADD COLUMN sender_type sender NOT NULL;

-- +goose Down
ALTER TABLE messages
DROP COLUMN sender_type IF EXISTS;

DROP TYPE IF EXISTS sender;
