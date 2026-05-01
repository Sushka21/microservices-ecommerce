-- +goose Up
-- +goose StatementBegin

create schema IF not exists loms;

create type loms.order_status as enum('new', 'awaiting payment', 'failed', 'paid', 'cancelled');

create table loms.orders
(
       id          bigint generated always as identity primary key,
       user_id     bigint            not null,
       status      loms.order_status not null default 'new',
       created_at  TIMESTAMPTZ       not null default now(),
       updated_at  TIMESTAMPTZ       not null default now()
);


create table loms.order_items
(
       order_id bigint not null references loms.orders (id) on delete cascade,
       sku      bigint not null,
       count    bigint not null check (count > 0),
       primary key (order_id, sku)
);

create or replace function loms.set_updated_at()
returns trigger as
$$
begin 
       new.updated_at = now();
       return new;
end
$$ language plpgsql;

create trigger trg_orders_set_updated_at
before update on loms.orders
for each row
execute function loms.set_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists loms.order_items;
drop table if exists loms.orders;
drop type if exists loms.order_status;

-- +goose StatementEnd
