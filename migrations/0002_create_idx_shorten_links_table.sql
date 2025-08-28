-- +goose Up
CREATE UNIQUE INDEX idx_shorten_links_original_url
    ON shorten_links (original_url);

-- +goose Down
DROP INDEX IF EXISTS idx_shorten_links_original_url;