-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;

-- name: GetUserFeeds :many
SELECT * FROM feeds
WHERE user_id = @user_id
ORDER BY id;

-- name: GetFeeds :many
SELECT feeds.*, users.name as user_name FROM feeds
JOIN users ON users.id = feeds.user_id
ORDER BY feeds.user_id;

-- name: GetFeedByUrl :one
SELECT * FROM feeds
where url = @url;