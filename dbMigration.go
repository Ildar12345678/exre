package main

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func migrateDB() {
	postgresDB, err := sql.Open("postgres", "user=gen dbname=expenses3")
	if err != nil {
		panic(err)
	}
	defer postgresDB.Close()

	stmtPostgres := `select p.purchase_date date, e.name name, s.name category, c.city, p.online, pc.count, pc.price from expense e
    left join subcat s on e.subcat_id = s.id
left outer join purchase_check pc on e.id = pc.expense_id
left join purchase p on pc.purchase_id = p.id
left join city c on p.city_id = c.id`

	rowsPostgres, err := postgresDB.Query(stmtPostgres)
	if err != nil {
		panic(err)
	}
	defer rowsPostgres.Close()

	sqliteDB, err := sql.Open("sqlite3", "./db/expenses.db")
	if err != nil {
		panic(err)
	}
	defer sqliteDB.Close()

	stmtSQLite := `create table if not exists expenses (
    id integer primary key autoincrement ,
    date date not null default current_date,
    name TEXT not null ,
    category text not null,
    city text not null ,
    online bool default false,
    count float4 not null,
    price float4 not null
);

create index if not exists idx_expenses_date on expenses(date);`

	_, err = sqliteDB.Exec(stmtSQLite)
	if err != nil {
		panic(err)
	}

	for rowsPostgres.Next() {
		var date time.Time
		var name string
		var category string
		var city string
		var online bool
		var count float32
		var price float32

		err = rowsPostgres.Scan(&date, &name, &category, &city, &online, &count, &price)
		if err != nil {
			panic(err)
		}
		stmtSQLite := `insert into expenses (date, name, category, city, online, count, price) values (?, ?, ?, ?, ?, ?, ?)`
		_, err = sqliteDB.Exec(stmtSQLite, date, name, category, city, online, count, price)
		if err != nil {
			panic(err)
		}
	}
}
