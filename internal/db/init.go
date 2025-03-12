package db

import (
	"expenses2/internal/db/sqlite"
	"expenses2/internal/db/mongo"
	"errors"
	"time"
	"expenses2/internal/types"
)

type DB interface {
	Initialize() error
	GetExpenses(time.Time, time.Time) ([]*types.ExpenseShow, error)
	GetStatistics(time.Time, time.Time) ([]*types.Statistics, error)
	SearchExpense(string) ([]*types.ExpenseSearch, error)
	GetCities() ([]string, error)
	GetCategories() ([]string, error)
	GetExpensesNames() ([]string, error)
	AddExpense(*types.ExpenseAdd) error
	UpdateExpense(*types.ExpenseShow) error
	DeleteExpense(int) error
	Close() error
}

func NewDB(dbType, dbPath, dbName string) (DB, error) {
	switch dbType {
	case "sqlite":
		return sqlite.NewSQLiteDB(dbPath, dbName)
	case "mongodb":
		return mongo.NewMongoDB(dbPath, dbName)
	default:
		return nil, errors.New("incorrect db choice")
	}
}
