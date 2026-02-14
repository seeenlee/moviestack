-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id           BIGSERIAL   NOT NULL PRIMARY KEY,
    username     TEXT        NOT NULL,
    display_name TEXT,
    bio          TEXT,
    avatar_url   TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS users_username_lower_unique ON users (lower(username));

-- +goose Down
DROP INDEX IF EXISTS users_username_lower_unique;
DROP TABLE IF EXISTS users;
