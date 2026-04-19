-- +goose Up
CREATE TABLE messages(
    id SERIAL PRIMARY KEY,
    conversation_id INT REFERENCES conversations (id) ON DELETE CASCADE,
    sender_id INT REFERENCES users (id) ON DELETE CASCADE,
    content TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    edited_at TIMESTAMP NOT NULL DEFAULT now()
);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION set_edited_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.edited_at  = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose  StatementEnd

-- +goose StatementBegin
CREATE TRIGGER trg_set_edited_at
BEFORE UPDATE ON messages
FOR EACH ROW
EXECUTE FUNCTION set_edited_at();
-- +goose  StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS trg_set_edited_at  ON messages;
DROP FUNCTION IF EXISTS set_edited_at;
DROP TABLE messages;