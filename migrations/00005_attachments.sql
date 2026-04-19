-- +goose Up
CREATE TABLE attachments(
    id SERIAL PRIMARY KEY,
    message_id INT NOT NULL REFERENCES messages (id) ON DELETE CASCADE,
    file_path TEXT NOT NULL,
    file_type VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- +goose Down
DROP TABLE attachments;
