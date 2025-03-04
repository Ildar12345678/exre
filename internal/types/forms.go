package types

import (
	"net/url"
	"strings"
)

// Formable interface is needed to transform data from form (after gin Bind function) to url.Values type
type Formable interface {
	Process() url.Values
}

type Form struct {
	Values url.Values
	Errors errors
}

func NewForm(data Formable) *Form {
	errs := make(map[string]string)
	return &Form{
		Values: data.Process(),
		Errors: errs,
	}
}

func (f *Form) Required(fields ...string) {
	for _, field := range fields {
		value := f.Values.Get(field)
		if strings.TrimSpace(value) == "" {
			f.Errors.Add(field, "This field cannot be blank")
		}
	}
}

func (f *Form) PermittedValues(field string, opts map[uint8]string) {
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

func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}
