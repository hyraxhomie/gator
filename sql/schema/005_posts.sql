-- +goose Up
CREATE TABLE posts(
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    title TEXT,
    url TEXT UNIQUE,
    description TEXT,
    published_at TIMESTAMP NOT NULL,
    feed_id UUID references feeds(id) ON DELETE CASCADE NOT NULL
);

-- +goose Down
DROP TABLE posts;