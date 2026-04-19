-- +goose Up
CREATE TYPE convo_type AS ENUM ('triad', 'duo');

CREATE TABLE conversations(
    id SERIAL PRIMARY KEY,
    conversation_type convo_type NOT NULL DEFAULT 'triad',
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE conversations;
DROP TYPE IF EXISTS convo_type;