-- +goose Up
alter table feeds 
ADD COLUMN last_fetched_at TIMESTAMP;

-- +goose Down
alter table feeds
drop COLUMN last_fetched_at;