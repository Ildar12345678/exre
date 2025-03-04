package app

import (
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
	GetExpensesNames(...any) ([]string, error)
	AddExpense(*types.ExpenseAdd) error
	Close() error
}
