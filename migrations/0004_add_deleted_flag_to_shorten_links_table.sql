-- +goose Up
ALTER TABLE shorten_links
    ADD COLUMN is_deleted BOOLEAN NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE shorten_links
    DROP COLUMN is_deleted;