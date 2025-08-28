-- +goose Up
ALTER TABLE shorten_links
    ADD COLUMN user_id TEXT NOT NULL DEFAULT '';

DROP INDEX IF EXISTS idx_shorten_links_original_url;

CREATE UNIQUE INDEX idx_shorten_links_original_url_user_id
    ON shorten_links (original_url, user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_shorten_links_original_url_user_id;

CREATE UNIQUE INDEX idx_shorten_links_original_url
    ON shorten_links (original_url);

ALTER TABLE shorten_links
    DROP COLUMN user_id;