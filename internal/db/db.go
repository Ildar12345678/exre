package dblayer

import (
	"fmt"
	"time"
	"database/sql"
	"expenses2/internal/types"
	"strings"
	"errors"
	_ "github.com/lib/pq"
)

type DB struct {
	db *sql.DB
}

func NewDB() (*DB, error) {
	dsn := "user=gen host=/var/run/postgresql dbname=expenses3"
	db, err := openDB(dsn)
	if err != nil {
		return nil, err
	}
	return &DB{db: db}, err
}

func (db *DB) GetExpenses(dateLow, dateHigh time.Time) ([]*types.ExpenseShow, error) {
	stmt := `select p.purchase_date as date,
                  round(c.count * c.price) as price,
                  e.name as expense,
                  s.name as subcat,
                  cat.name as cat
						from purchase as p, purchase_check as c, expense as e, subcat as s, cat
						where e.subcat_id = s.id and
									c.expense_id = e.id and
									c.purchase_id = p.id and
									s.cat_id = cat.id and
									p.purchase_date >= $1 and
									p.purchase_date <= $2 order by date`
	
	rows, err := db.db.Query(stmt, dateLow, dateHigh)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	dest := make([]*types.ExpenseShow, 0, 100)
	for rows.Next() {
		var expense types.ExpenseShow
		err = rows.Scan(&expense.Date, &expense.Price, &expense.Expense, &expense.Subcat, &expense.Cat)
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

func (db *DB) GetStatistics(dateLow, dateHigh time.Time) ([]*types.Statistics, error) {
	stmt := `select s.name subcat, cat.name cat, round(sum(c.count * c.price)) sum_subcat
					 from subcat s, cat, purchase p, expense e
           left outer join purchase_check c on e.id = c.expense_id
					 where e.subcat_id = s.id and
						 cat.id = s.cat_id and
             p.id = c.purchase_id and
             p.purchase_date >= $1 and
						 p.purchase_date <= $2
					 group by s.name, cat.name order by sum_subcat desc `
	
	rows, err := db.db.Query(stmt, dateLow, dateHigh)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	dest := make([]*types.Statistics, 0, 100)
	for rows.Next() {
		var stat types.Statistics
		err = rows.Scan(&stat.Subcat, &stat.Cat, &stat.SumSubcat)
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

// SearchExpense can find all expenses with certain name with no case sensitivity
func (db *DB) SearchExpense(name string) ([]*types.ExpenseSearch, error) {
	stmt := `select e.name, c.price, p.purchase_date date
from expense e, purchase_check c, purchase p
where e.id = c.expense_id and p.id = c.purchase_id and e.name like '%'||$1||'%' order by purchase_date`
	
	dest := make([]*types.ExpenseSearch, 0, 100)
	
	for _, s := range []string{name, strings.ToUpper(name), strings.ToLower(name)} {
		rows, err := db.db.Query(stmt, s)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var expense types.ExpenseSearch
			rows.Scan(&expense)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return nil, types.ErrNoRecord
				} else {
					return nil, err
				}
			}
			dest = append(dest, &expense)
		}
	}
	
	return dest, nil
}

func (db *DB) GetCity() ([]*types.City, error) {
	stmt := "select id, city from city"
	rows, err := db.db.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	dest := make([]*types.City, 0, 10)
	for rows.Next() {
		var city types.City
		err = rows.Scan(&city.ID, &city.City)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, types.ErrNoRecord
			} else {
				return nil, err
			}
		}
		dest = append(dest, &city)
	}
	return dest, nil
}

// func (db *DB) GetCat() ([]*types.Cat, error) {
// 	stmt := "select id, name from cat"
// 	rows, err := db.db.Query(stmt)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()
// 	dest := make([]*types.Cat, 0, 10)
// 	for rows.Next() {
// 		var cat types.Cat
// 		rows.Scan(&cat)
// 		dest = append(dest, &cat)
// 	}
//
// 	return dest, nil
// }

func (db *DB) GetSubcat() ([]*types.Subcat, error) {
	stmt := "select id, name, cat_id from subcat order by cat_id"
	rows, err := db.db.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	dest := make([]*types.Subcat, 0, 10)
	for rows.Next() {
		var subcat types.Subcat
		err = rows.Scan(&subcat.ID, &subcat.Name, &subcat.Cat_id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, types.ErrNoRecord
			} else {
				return nil, err
			}
		}
		dest = append(dest, &subcat)
	}
	return dest, nil
}

// func (db *DB) GetSupplier() ([]*types.MOS, error) {
// 	stmt := "select name, address, id from market_or_supplier"
// 	rows, err := db.db.Query(stmt)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()
// 	dest := make([]*types.MOS, 0, 10)
// 	for rows.Next() {
// 		var mos types.MOS
// 		rows.Scan(&mos)
// 		dest = append(dest, &mos)
// 	}
//
// 	return dest, nil
// }

func (db *DB) GetExpensesNames() ([]*types.Expense, error) {
	stmt := "select id, name, subcat_id, nds from expense"
	rows, err := db.db.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	dest := make([]*types.Expense, 0, 100)
	for rows.Next() {
		var expense types.Expense
		err = rows.Scan(&expense.ID, &expense.Name, &expense.Subcat_id, &expense.NDS)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, types.ErrNoRecord
			} else {
				return nil, err
			}
		}
		dest = append(dest, &expense)
	}
	return dest, nil
}

func (db *DB) AddExpense(expenses *types.ExpenseAdd) error {
	var row *sql.Rows
	var err error
	var stmt string
	
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func closeDB(db *DB) {
	if err := db.db.Close(); err != nil {
		fmt.Println("error while closing DB:", err)
	}
}
