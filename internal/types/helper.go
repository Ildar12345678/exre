package types

import (
	"fmt"
	"mime/multipart"
	"net/url"
)

func (e *ExpenseAdd) Process() (url.Values, *multipart.FileHeader) {
	vals := map[string][]string{
		"date":     {e.Date},
		"name":     {e.Name},
		"category": {e.Category},
		"city":     {e.City},
		"online":   {fmt.Sprintf("%t", e.Online)},
		"count":    {e.Count},
		"price":    {e.Price},
	}
	return vals, nil
}

func (e *ExpenseAddFromCheck) Process() (url.Values, *multipart.FileHeader) {
	vals := map[string][]string{
		"city":     {e.City},
		"category": {e.Category},
	}
	return vals, e.File
}
