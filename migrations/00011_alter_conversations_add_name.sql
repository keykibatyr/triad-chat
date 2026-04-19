-- +goose Up
ALTER TABLE conversations 
ADD COLUMN name TEXT NOT NULL;

-- +goose Down
ALTER TABLE conversations 
DROP COLUMN name IF EXISTS;
