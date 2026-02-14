-- name: ListUsers :many
SELECT id, username, display_name, bio, avatar_url, created_at, updated_at
FROM users
ORDER BY id DESC;

-- name: CreateUser :one
INSERT INTO users (username)
VALUES (@username)
RETURNING id, username, display_name, bio, avatar_url, created_at, updated_at;

-- name: DeleteUser :execrows
DELETE FROM users
WHERE id = @id;

-- name: UserExists :one
SELECT EXISTS (
    SELECT 1
    FROM users
    WHERE id = @id
);
