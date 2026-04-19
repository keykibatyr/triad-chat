-- +goose Up
CREATE TYPE message_type AS ENUM 
('message', 'join_room',
 'leave_room', 'system',
  'ai_message');

ALTER TABLE messages
ADD COLUMN type message_type NOT NULL;

-- +goose Down
ALTER TABLE messages
DROP COLUMN type IF EXISTS;

DROP TYPE IF EXISTS message_type;
