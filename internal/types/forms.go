package types

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/url"
	"strings"
)

// Formable interface is needed to transf data from f (after gin Bind function) to url.Values type
type Formable interface {
	Process() (url.Values, *multipart.FileHeader)
}

type FormAddExpense struct {
	Multipart *multipart.FileHeader
	Values url.Values
	Errors errors
}

func NewForm(data Formable) *FormAddExpense {
	errs := make(map[string]string)
	urlValues, file := data.Process()
	return &FormAddExpense{
		Values: urlValues,
		Multipart: file,
		Errors: errs,
	}
}

func (f *FormAddExpense) Required(fields ...string) {
	for _, field := range fields {
		value := f.Values.Get(field)
		if strings.TrimSpace(value) == "" {
			f.Errors.Add(field, "This field cannot be blank")
		}
	}
}

func (f *FormAddExpense) CheckFile()  (*Check, error) {
	file, err := f.Multipart.Open()
	if err != nil {
		f.Errors.Add("file", "can't open the file")
		return nil, err
	}
	
	defer file.Close()
	fullCheck := make([]MainDoc, 0)
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&fullCheck)
	if err != nil {
		f.Errors.Add("file", "error while reading the file")
		return nil, err
	}
	if len(fullCheck) == 0 {
		f.Errors.Add("file", "incorrect data in the file")
		return nil, fmt.Errorf("incorrect data in the file")
	}
	check := fullCheck[0].Ticket.Document.Receipt
	splittedDate := strings.Split(check.Date, "T")
	if len(splittedDate) == 0 {
		f.Errors.Add("file", "incorrect date in the file")
		return nil, fmt.Errorf("incorrect date in the file")
	}
	check.Date = splittedDate[0]

	return &check, nil
}

func (f *FormAddExpense) PermittedValues(field string, opts map[uint8]string) {
	value := f.Values.Get(field)
	if value == "" {
		return
	}
	for _, opt := range opts {
		if value == opt {
			return
		}
	}
	f.Errors.Add(field, "This value is invalid")
}

func (f *FormAddExpense) Valid() bool {
	return len(f.Errors) == 0
}
