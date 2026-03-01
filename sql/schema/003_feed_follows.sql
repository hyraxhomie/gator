-- +goose Up
CREATE TABLE feed_follows(
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    user_id UUID references users(id) ON DELETE CASCADE NOT NULL,
    feed_id UUID references feeds(id) ON DELETE CASCADE NOT NULL
);

-- +goose Down
DROP TABLE feed_follows;