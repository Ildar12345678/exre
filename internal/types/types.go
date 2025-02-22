package types

import (
	"time"
	"database/sql"
	"errors"
)

var SubCategories = map[uint8]string{
	// здоровье
	1: "лекарство", 2: "обследования", 3: "уход за собой", 4: "профилактика",
	// непродукты
	5: "одежда", 6: "обувь", 7: "аксессуары", 8: "мебель", 9: "электроника", 10: "хозтовары",
	11: "для ремонта", 12: "книги", 13: "спорт", 14: "развлечения", 15: "косметика", 16: "другое",
	// продукты
	17: "для дома", 18: "фр/сфр/ор", 19: "неполезное", 20: " не дома", 35: "животным",
	// проезд
	21: "межгород", 22: "такси", 23: "общественный",
	// развлеченья
	24: "кафе", 25: "культуры", 26: "заказ еды", 27: "путешествия", 28: "другое",
	// другое
	29: "услуги", 30: "государству", 31: "подарки", 32: "сотовый", 33: "благотвор", 34: "другое",
}

var ErrNoRecord = errors.New("no matching record found")

type Subcat struct {
	ID     int    `db:"id"`
	Name   string `db:"name"`
	Cat_id int    `db:"cat_id"`
}

type City struct {
	ID   int    `db:"id"`
	City string `db:"city"`
}

type Expense struct {
	ID        int           `db:"id"`
	Name      string        `db:"name"`
	Subcat_id int           `db:"subcat_id"`
	NDS       sql.NullInt32 `db:"nds"`
}

type Reply struct {
	Id          int            `db:"id"`
	Rate        int            `db:"rate"`
	Description sql.NullString `db:"description"`
	Mos_id      sql.NullInt32  `db:"mos_id"`
	Expense_id  sql.NullInt32  `db:"expense_id"`
}

type Purchase struct {
	ID            int            `db:"id"`
	Purchase_date time.Time      `db:"purchase_date"`
	City_id       int            `db:"city_id"`
	Online        bool           `db:"online"`
	Description   sql.NullString `db:"description"`
	Mos_id        sql.NullInt32  `db:"mos_id"`
}

type PurchaseCheck struct {
	ID          int     `db:"id"`
	Purchase_id int     `db:"purchase_id"`
	Expense_id  int     `db:"expense_id"`
	Count       float32 `db:"count"`
	Price       int     `db:"price"`
}

type TemplateResult struct {
	URL          string
	ExpensesShow []*ExpenseShow
	Statistics   StatAndSum
	AddShow      AddShowForm
}

type AddShowForm struct {
	Date        string
	Cities      []*City
	ExpenseName []*Expense
	Subcat      []*Subcat
	Online      []string
	Nds         []string
}

type FilterExpenses struct {
	DateLow  string `form:"date_low"`
	DateHigh string `form:"date_high"`
	Subcat   string `form:"subcat"`
}

type ExpenseAdd struct {
	Date   string `form:"date"`
	Name   string `form:"name"`
	Subcat string `form:"subcat"`
	City   string `form:"city"`
	Online bool   `form:"online"`
	Count  string `form:"count"`
	Price  string `form:"price"`
	NDS    int    `form:"nds"`
}

type StatAndSum struct {
	Sum        int
	Statistics []*Statistics
	PngSubcat  [][]string
	PngCat     [][]string
}

type ExpenseShow struct {
	Date    string  `db:"date"`
	Price   float32 `db:"price"`
	Expense string  `db:"expense"`
	Subcat  string  `db:"subcat"`
	Cat     string  `db:"cat"`
}

type Statistics struct {
	Cat       string  `db:"cat"`
	Subcat    string  `db:"subcat"`
	SumSubcat float32 `db:"sum_subcat"`
}

type ExpenseSearch struct {
	Name  string    `db:"name"`
	Price string    `json:"price"`
	Date  time.Time `json:"date"`
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
	Subcat   uint8
}
