-- +goose Up
CREATE TYPE user_role AS ENUM ('admin', 'user');

CREATE TABLE users(
    id SERIAL PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role user_role NOT NULL DEFAULT 'user',
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT  NULL DEFAULT now()
);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose  StatementEnd

-- +goose  StatementBegin
CREATE TRIGGER trg_set_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS trg_set_updated_at ON users;
DROP FUNCTION IF EXISTS set_updated_at;
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS user_role;