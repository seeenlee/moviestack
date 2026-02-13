-- name: SearchMovies :many
SELECT id, original_title, adult, video, popularity,
       similarity(original_title, @query) AS score
FROM movie_ids
WHERE similarity(original_title, @query) > 0.1
ORDER BY score DESC, popularity DESC
LIMIT 20;
