CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX idx_movie_ids_title_trgm ON movie_ids USING GIN (original_title gin_trgm_ops);
