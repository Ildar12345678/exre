package types

import (
	"net/url"
	"fmt"
)

func (e *ExpenseAdd) Process() url.Values {
	vals := map[string][]string{
		"date":   {e.Date},
		"name":   {e.Name},
		"subcat": {e.Subcat},
		"city":   {e.City},
		"online": {fmt.Sprintf("%t", e.Online)},
		"count":  {e.Count},
		"price":  {e.Price},
		"nds":    {e.NDS},
	}
	return vals
}
