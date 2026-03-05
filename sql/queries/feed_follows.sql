-- name: CreateFeedFollow :one
WITH feed_follow as (INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *)
SELECT feed_follow.*,
users.name as user_name,
feeds.name as feed_name
from feed_follow
join users on users.id = feed_follow.user_id
join feeds on feeds.id = feed_follow.feed_id;

-- name: GetFeedFollowsForUser :many
SELECT feed_follows.*, feeds.name as feed_name
from feed_follows
join feeds on feeds.id = feed_follows.feed_id
where feed_follows.user_id = @user_id;

-- name: DeleteFeedFollow :one
with deleted as (
    delete from feed_follows
    using feeds
    where feeds.id = feed_follows.feed_id
    and feeds.url = @url
    and feed_follows.user_id = @user_id
    RETURNING feed_follows.feed_id
)
select feeds.*
from deleted 
join feeds on feeds.id = deleted.feed_id;