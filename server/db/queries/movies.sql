-- name: SearchMovies :many
SELECT id, original_title, adult, video, popularity,
       similarity(original_title, @query) AS score
FROM movie_ids
WHERE original_title % @query
ORDER BY score DESC, popularity DESC
LIMIT 20;

-- name: MovieExists :one
SELECT EXISTS (
    SELECT 1
    FROM movie_ids
    WHERE id = @id
);
