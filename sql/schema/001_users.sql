-- +goose Up
-- SQL to create the users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    email TEXT NOT NULL UNIQUE
);

-- +goose Down
-- SQL to drop the users table
DROP TABLE IF EXISTS users;
