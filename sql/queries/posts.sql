-- name: CreatePost :one
INSERT INTO posts (id, created_at, updated_at, title, url, description, published_at, feed_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
)
RETURNING *;

-- name: GetPostsForUser :many
select posts.* from posts
join feed_follows on feed_follows.feed_id = posts.feed_id
where feed_follows.user_id = @user_id
ORDER BY posts.published_at DESC LIMIT @num_posts::int;