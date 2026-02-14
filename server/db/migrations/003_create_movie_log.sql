-- +goose Up
CREATE TABLE IF NOT EXISTS movie_log (
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

CREATE UNIQUE INDEX IF NOT EXISTS movie_log_user_rank_unique
    ON movie_log (user_id, rank_position)
    WHERE rank_position IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_movie_log_user_watched_on ON movie_log (user_id, watched_on DESC);
CREATE INDEX IF NOT EXISTS idx_movie_log_user_created_at ON movie_log (user_id, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_movie_log_user_created_at;
DROP INDEX IF EXISTS idx_movie_log_user_watched_on;
DROP INDEX IF EXISTS movie_log_user_rank_unique;
DROP TABLE IF EXISTS movie_log;
