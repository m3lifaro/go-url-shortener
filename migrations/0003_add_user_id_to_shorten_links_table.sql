-- +goose Up
ALTER TABLE shorten_links
    ADD COLUMN user_id TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE shorten_links
    DROP COLUMN user_id;