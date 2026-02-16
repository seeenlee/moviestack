-- +goose Up
CREATE TABLE IF NOT EXISTS movie_ids (
    id             INTEGER        NOT NULL PRIMARY KEY,
    original_title TEXT           NOT NULL,
    adult          BOOLEAN        NOT NULL DEFAULT false,
    video          BOOLEAN        NOT NULL DEFAULT false,
    popularity     NUMERIC(10, 4) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_movie_ids_popularity ON movie_ids (popularity);

-- +goose Down
DROP INDEX IF EXISTS idx_movie_ids_popularity;
DROP TABLE IF EXISTS movie_ids;
