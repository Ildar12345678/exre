package types

import (
	"fmt"
	"mime/multipart"
)

var ErrNoRecord = fmt.Errorf("no matching record found")

type TemplateResult struct {
	URL           string
	DateLow       string
	DateHigh      string
	SearchPattern string
	ExpensesShow  []*ExpenseShow
	Statistics    StatAndSum
	AddShow       AddExpenseShowForm
	SearchResult  []*ExpenseSearch
}

type StatAndSum struct {
	Sum         int
	Statistics  []*Statistics
	PngCategory [][4]string
}

type AddExpenseShowForm struct {
	Date        string
	Name        []string
	Category    []string
	DefaultCity string
	Cities      []string
	Online      []string
	Form        *FormAddExpense
}

type FilterExpenses struct {
	DateLow  string `form:"date_low"`
	DateHigh string `form:"date_high"`
	URL      string `form:"url"`
	Category string `form:"category"`
}

type ExpenseAdd struct {
	Date     string `form:"date"`
	Name     string `form:"name"`
	Category string `form:"category"`
	City     string `form:"city"`
	Online   bool   `form:"online"`
	Count    string `form:"count"`
	Price    string `form:"price"`
}

type ExpenseAddFromCheck struct {
	City     string
	Category string
	File     *multipart.FileHeader
}

type ExpenseShow struct {
	ID       int     `db:"id"`
	Date     string  `db:"date"`
	Price    float32 `db:"price"`
	Name     string  `db:"name"`
	Category string  `db:"category"`
}

type Statistics struct {
	Category    string  `db:"cat"`
	SumCategory float32 `db:"sum_category"`
}

type ExpenseSearch struct {
	Name  string `db:"name"`
	Price string `json:"price"`
	Date  string `json:"date"`
}

// all types below refer to check parse

type MainDoc struct {
	Ticket Ticket `json:"ticket"`
}

type Ticket struct {
	Document Document `json:"document"`
}

type Document struct {
	Receipt Check `json:"receipt"`
}

type Check struct {
	Date  string `json:"dateTime"`
	Items []Item `json:"items"`
	City  string
}

type Item struct {
	Name     string  `json:"name"`
	Nds      int     `json:"nds"`   // 1 -> 20%, 2 -> 10%
	Price    float32 `json:"price"` // need to divide by 100
	Quantity float32 `json:"quantity"`
	Category string
}
