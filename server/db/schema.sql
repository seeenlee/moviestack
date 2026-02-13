CREATE TABLE movie_ids (
    id             INTEGER        NOT NULL PRIMARY KEY,
    original_title TEXT           NOT NULL,
    adult          BOOLEAN        NOT NULL DEFAULT false,
    video          BOOLEAN        NOT NULL DEFAULT false,
    popularity     NUMERIC(10, 4) NOT NULL
);

CREATE INDEX idx_movie_ids_popularity ON movie_ids (popularity);
