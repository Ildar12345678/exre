package sqlite

import (
	"database/sql"
	"errors"
	"expenses2/internal/types"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteDB struct {
	conn   *sql.DB
	dbPath string
}

func NewSQLiteDB(dbPath, dbName string) (*SQLiteDB, error) {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		if err = os.Mkdir(dbPath, 0775); err != nil {
			return nil, fmt.Errorf("error while creating db dir: %s", err.Error())
		}
	}

	conn, err := sql.Open("sqlite3", filepath.Join(dbPath, dbName))
	if err != nil {
		return nil, err
	}
	return &SQLiteDB{conn: conn, dbPath: dbPath}, nil
}

func (d *SQLiteDB) Initialize() error {
	// Initialize database schema
	initSQL := `create table if not exists expenses (
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

	_, err := d.conn.Exec(initSQL)
	if err != nil {
		return fmt.Errorf("unable to initialize DB: %s", err.Error())
	}
	return nil
}

func (d *SQLiteDB) GetExpenses(dateLow, dateHigh time.Time) ([]*types.ExpenseShow, error) {
	stmt := `select id, date, round(count * price) as price, name, category
						from expenses where date >= ? and date <= ? order by date`
	rows, err := d.conn.Query(stmt, dateLow.Format("2006-01-02"), dateHigh.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	dest := make([]*types.ExpenseShow, 0, 100)
	for rows.Next() {
		var expense types.ExpenseShow
		err = rows.Scan(&expense.ID, &expense.Date, &expense.Price, &expense.Name, &expense.Category)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, types.ErrNoRecord
			} else {
				return nil, err
			}
		}
		expense.Date = strings.Split(expense.Date, "T")[0]
		dest = append(dest, &expense)
	}
	return dest, nil
}

func (d *SQLiteDB) GetStatistics(dateLow, dateHigh time.Time) ([]*types.Statistics, error) {
	stmt := `select category, round(sum(count * price)) sum_category
from expenses where date >= ? and date <= ? group by category order by category desc`

	rows, err := d.conn.Query(stmt, dateLow.Format("2006-01-02"), dateHigh.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	dest := make([]*types.Statistics, 0, 100)
	for rows.Next() {
		var stat types.Statistics
		err = rows.Scan(&stat.Category, &stat.SumCategory)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, types.ErrNoRecord
			} else {
				return nil, err
			}
		}
		dest = append(dest, &stat)
	}
	return dest, nil
}

func (d *SQLiteDB) SearchExpense(name string) ([]*types.ExpenseSearch, error) {
	stmt := `select name, price, date from expenses where name like '%'||?||'%' order by date`

	dest := make([]*types.ExpenseSearch, 0, 100)

	rows, err := d.conn.Query(stmt, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var expense types.ExpenseSearch
		err = rows.Scan(&expense.Name, &expense.Price, &expense.Date)
		if err != nil {
			return nil, err
		}
		expense.Date = strings.Split(expense.Date, "T")[0]
		dest = append(dest, &expense)
	}
	if len(dest) == 0 {
		return nil, types.ErrNoRecord
	}
	return dest, nil
}

func (d *SQLiteDB) GetCities() ([]string, error) {
	stmt := "select distinct city from expenses"
	rows, err := d.conn.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	dest := make([]string, 0, 100)
	for rows.Next() {
		var city string
		err = rows.Scan(&city)
		if err != nil {
			return nil, err
		}
		dest = append(dest, city)
	}
	return dest, nil
}

func (d *SQLiteDB) GetCategories() ([]string, error) {
	stmt := "select distinct category from expenses"
	rows, err := d.conn.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	dest := make([]string, 0, 100)
	for rows.Next() {
		var category string
		err = rows.Scan(&category)
		if err != nil {
			return nil, err
		}
		dest = append(dest, category)
	}
	return dest, nil
}

func (d *SQLiteDB) GetExpensesNames() ([]string, error) {
	stmt := "select distinct name from expenses"

	rows, err := d.conn.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	dest := make([]string, 0, 100)
	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, types.ErrNoRecord
			} else {
				return nil, err
			}
		}
		dest = append(dest, name)
	}
	return dest, nil
}

func (d *SQLiteDB) AddExpense(expense *types.ExpenseAdd) error {
	stmt := `insert into expenses (date,name,category,city,online,count,price) values (?,?,?,?,?,?,?)`
	_, err := d.conn.Exec(stmt, expense.Date, expense.Name, expense.Category, expense.City,
		expense.Online, expense.Count, expense.Price)
	if err != nil {
		return fmt.Errorf("expense insert error: %v", err)
	}
	return nil
}

func (d *SQLiteDB) UpdateExpense(expense *types.ExpenseShow) error {
	stmt := `update expenses set date=?,name=?,category=? where id=?`
	_, err := d.conn.Exec(stmt, expense.Date, expense.Name, expense.Category, expense.ID)
	if err != nil {
		return fmt.Errorf("expense update error: %v", err)
	}
	return nil
}

func (d *SQLiteDB) DeleteExpense(id int) error {
	stmt := `delete from expenses where id = ?`
	_, err := d.conn.Exec(stmt, id)
	if err != nil {
		return err
	}
	return nil
}

func (d *SQLiteDB) Close() error {
	return d.conn.Close()
}
