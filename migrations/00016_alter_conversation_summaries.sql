-- +goose Up

DROP TRIGGER IF EXISTS trg_set_edited_at ON conversation_summaries;

ALTER TABLE conversation_summaries
ALTER COLUMN conversation_id SET NOT NULL;

ALTER TABLE conversation_summaries
ALTER COLUMN summary_text SET DEFAULT '';

UPDATE conversation_summaries
SET summary_text = ''
WHERE summary_text IS NULL;

ALTER TABLE conversation_summaries
ALTER COLUMN summary_text SET NOT NULL;

ALTER TABLE conversation_summaries
DROP CONSTRAINT IF EXISTS conversation_summaries_last_message_id_fkey;

ALTER TABLE conversation_summaries
ADD CONSTRAINT conversation_summaries_conversation_id_unique UNIQUE (conversation_id);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER trg_conversation_summaries_updated_at
BEFORE UPDATE ON conversation_summaries
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- +goose Down

DROP TRIGGER IF EXISTS trg_conversation_summaries_updated_at ON conversation_summaries;
DROP FUNCTION IF EXISTS set_updated_at;

ALTER TABLE conversation_summaries
DROP CONSTRAINT IF EXISTS conversation_summaries_conversation_id_unique;

ALTER TABLE conversation_summaries
ADD CONSTRAINT conversation_summaries_last_message_id_fkey
FOREIGN KEY (last_message_id) REFERENCES messages(id) ON DELETE CASCADE;

ALTER TABLE conversation_summaries
ALTER COLUMN summary_text DROP NOT NULL;

ALTER TABLE conversation_summaries
ALTER COLUMN conversation_id DROP NOT NULL;