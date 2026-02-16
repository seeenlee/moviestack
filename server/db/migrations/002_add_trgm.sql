-- +goose Up
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX IF NOT EXISTS idx_movie_ids_title_trgm
    ON movie_ids USING GIN (original_title gin_trgm_ops);

-- +goose Down
DROP INDEX IF EXISTS idx_movie_ids_title_trgm;
DROP EXTENSION IF EXISTS pg_trgm;
