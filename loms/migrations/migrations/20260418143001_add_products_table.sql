-- +goose Up
-- +goose StatementBegin

create table loms.products
(
    sku   integer generated always as identity primary key,
    names text    not null,
    price integer not null check (price > 0)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

drop table if exists loms.products;

-- +goose StatementEnd