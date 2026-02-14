CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE movie_ids (
    id             INTEGER        NOT NULL PRIMARY KEY,
    original_title TEXT           NOT NULL,
    adult          BOOLEAN        NOT NULL DEFAULT false,
    video          BOOLEAN        NOT NULL DEFAULT false,
    popularity     NUMERIC(10, 4) NOT NULL
);

CREATE INDEX idx_movie_ids_popularity ON movie_ids (popularity);
CREATE INDEX idx_movie_ids_title_trgm ON movie_ids USING GIN (original_title gin_trgm_ops);

CREATE TABLE users (
    id           BIGSERIAL   NOT NULL PRIMARY KEY,
    username     TEXT        NOT NULL,
    display_name TEXT,
    bio          TEXT,
    avatar_url   TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX users_username_lower_unique ON users (lower(username));
