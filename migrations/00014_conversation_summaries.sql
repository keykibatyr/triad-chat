-- +goose Up
CREATE TABLE conversation_summaries (
    id SERIAL PRIMARY KEY,
    conversation_id INT REFERENCES conversations (id) ON DELETE CASCADE,
    summary_text TEXT,
    last_message_id INT REFERENCES messages (id) ON DELETE CASCADE, 
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION set_edited_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at  = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose  StatementEnd

-- +goose StatementBegin
CREATE TRIGGER trg_set_edited_at
BEFORE UPDATE ON conversation_summaries
FOR EACH ROW
EXECUTE FUNCTION set_edited_at();
-- +goose  StatementEnd


-- +goose Down
DROP TRIGGER IF EXISTS trg_set_edited_at ON conversation_summaries;
DROP FUNCTION IF EXISTS set_edited_at;
DROP TABLE conversation_summaries;