-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE lower(name) = lower(@name)
ORDER BY id;

-- name: GetUsers :many
SELECT * FROM users
ORDER BY id;

-- name: DeleteUsers :exec
DELETE FROM users;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = @id;