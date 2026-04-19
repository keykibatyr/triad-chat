-- +goose Up
CREATE TABLE message_status(
    message_id INT REFERENCES messages (id) ON DELETE CASCADE,
    user_id INT REFERENCES users (id) ON DELETE CASCADE,
    delivered_at TIMESTAMP NULL,
    read_at TIMESTAMP NULL,
    PRIMARY KEY (message_id, user_id)
);
-- +goose Down
DROP TABLE message_status;