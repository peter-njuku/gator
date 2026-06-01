-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name, password_hash)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
where Name = $1 limit 1;

-- name: DeleteAllUsers :exec
DELETE FROM users;

-- name: GetUsers :many
SELECT * FROM users;

-- name: GetUserById :one
SELECT * FROM users WHERE id = $1;