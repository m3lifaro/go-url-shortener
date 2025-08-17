-- +goose Up
create table if not exists shorten_links (
    id serial primary key,
    short_url text not null ,
    original_url text not null
);

-- +goose Down
drop table if exists shorten_links;