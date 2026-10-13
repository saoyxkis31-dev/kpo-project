alter table products
add column stock integer not null default 0 check (stock >= 0);