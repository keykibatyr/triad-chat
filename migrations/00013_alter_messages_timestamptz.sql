-- +goose Up
ALTER TABLE messages
    ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC',
    ALTER COLUMN edited_at TYPE TIMESTAMPTZ USING edited_at AT TIME ZONE 'UTC';

ALTER TABLE messages
    ALTER COLUMN created_at SET DEFAULT now(),
    ALTER COLUMN edited_at SET DEFAULT now();

-- +goose Down
ALTER TABLE messages
    ALTER COLUMN created_at TYPE TIMESTAMP USING created_at AT TIME ZONE 'UTC',
    ALTER COLUMN edited_at TYPE TIMESTAMP USING edited_at AT TIME ZONE 'UTC';

ALTER TABLE messages
    ALTER COLUMN created_at SET DEFAULT now(),
    ALTER COLUMN edited_at SET DEFAULT now();