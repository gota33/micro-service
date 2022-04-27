create table if not exists item
(
    id          integer primary key autoincrement,
    title       text           not null check (length(title) > 0),
    price       decimal(12, 2) not null check (price >= 0),
    num         integer        not null check (num >= 0),
    create_time timestamp      not null default current_timestamp
);

create table if not exists customer
(
    id          integer primary key autoincrement,
    nick        text           not null check (length(nick) > 0),
    balance     decimal(12, 2) not null check (balance >= 0),
    create_time timestamp      not null default current_timestamp
);

create table if not exists cart
(
    id          integer primary key autoincrement,
    customer_id integer   not null,
    item_id     integer   null,
    num         integer   not null check (num >= 0),
    status      text      not null check (status IN ('open', 'closed')),
    create_time timestamp not null default current_timestamp,
    foreign key (customer_id) references customer (id) on delete cascade,
    foreign key (item_id) references item (id) on delete set null
);

create index if not exists idx_cart on cart (customer_id, status);
