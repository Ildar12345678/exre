create table if not exists expenses (
    id integer primary key autoincrement ,
    date date not null default current_date,
    name TEXT not null ,
    category text not null,
    city text not null ,
    online bool default false,
    count float4 not null,
    price float4 not null
);

create index idx_expenses_date on expenses(date);