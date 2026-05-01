-- +goose Up
-- +goose StatementBegin

create schema IF not exists cart;

create table cart.items 
(
       user_id bigint not null,
       sku     bigint not null,
       count   bigint not null check (count > 0),
       primary key (user_id, sku)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

drop table if exists cart.items;

-- +goose StatementEnd