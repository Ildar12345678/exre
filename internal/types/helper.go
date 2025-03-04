package types

import (
	"net/url"
	"fmt"
)

func (e *ExpenseAdd) Process() url.Values {
	vals := map[string][]string{
		"date":     {e.Date},
		"name":     {e.Name},
		"category": {e.Category},
		"city":     {e.City},
		"online":   {fmt.Sprintf("%t", e.Online)},
		"count":    {fmt.Sprintf("%f", e.Count)},
		"price":    {e.Price},
	}
	return vals
}
