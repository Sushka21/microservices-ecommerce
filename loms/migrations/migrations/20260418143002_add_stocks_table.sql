-- +goose Up
-- +goose StatementBegin

create table loms.available_stocks
(
      sku   integer not null primary key references loms.products(sku) on delete cascade,
      count integer not null check (count >= 0)
);

create table loms.reserved_stocks
(
      sku      integer not null references loms.products(sku) on delete cascade,
      order_id bigint  not null references loms.orders (id) on delete cascade,
      count    integer not null check (count >= 0),    
      primary key (sku, order_id)
);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

drop table if exists loms.reserved_stocks;
drop table if exists loms.available_stocks;

-- +goose StatementEnd