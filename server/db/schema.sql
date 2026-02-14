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

CREATE TABLE movie_log (
    id            BIGSERIAL   NOT NULL PRIMARY KEY,
    user_id       BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    movie_id      INTEGER     NOT NULL REFERENCES movie_ids (id) ON DELETE CASCADE,
    watched_on    DATE        NOT NULL DEFAULT CURRENT_DATE,
    note          TEXT,
    rank_position INTEGER,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT movie_log_user_movie_unique UNIQUE (user_id, movie_id),
    CONSTRAINT movie_log_rank_positive CHECK (rank_position IS NULL OR rank_position > 0)
);

CREATE UNIQUE INDEX movie_log_user_rank_unique
    ON movie_log (user_id, rank_position)
    WHERE rank_position IS NOT NULL;
CREATE INDEX idx_movie_log_user_watched_on ON movie_log (user_id, watched_on DESC);
CREATE INDEX idx_movie_log_user_created_at ON movie_log (user_id, created_at DESC);
