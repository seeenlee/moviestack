-- name: ListMovieLogByUser :many
SELECT ml.id AS log_id, ml.user_id, ml.movie_id, mi.original_title, ml.watched_on,
       ml.note, ml.rank_position, ml.created_at, ml.updated_at
FROM movie_log ml
JOIN movie_ids mi ON mi.id = ml.movie_id
WHERE ml.user_id = @user_id
ORDER BY (ml.rank_position IS NULL), ml.rank_position ASC NULLS LAST, ml.created_at DESC;

-- name: UpsertMovieLogEntry :one
INSERT INTO movie_log (user_id, movie_id, watched_on, note, rank_position)
VALUES (@user_id, @movie_id, @watched_on, @note, NULL)
ON CONFLICT (user_id, movie_id) DO UPDATE
SET watched_on = EXCLUDED.watched_on,
    note = EXCLUDED.note,
    updated_at = now()
RETURNING id, user_id, movie_id, watched_on, note, rank_position, created_at, updated_at;

-- name: DeleteMovieLogEntry :execrows
DELETE FROM movie_log
WHERE id = @id AND user_id = @user_id;
