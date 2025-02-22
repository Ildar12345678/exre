package app

import (
	"net/http"
	"fmt"
	"runtime/debug"
	"time"
	"math"
	"expenses2/internal/types"
)

// The serverError helper writes an error message and stack trace to the errorLog,
// then sends a generic 500 Internal Server Error response to the user.
func (a *App) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	a.logger.Errorf(trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// The clientError helper sends a specific status code and corresponding description
// to the user. We'll use this later in the book to send responses like 400 "Bad
// Request" when there's a problem with the request that the user sent.
func (a *App) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// For consistency, we'll also implement a notFound helper. This is simply a
// convenience wrapper around clientError which sends a 404 Not Found response to
// the user.
func (a *App) notFound(w http.ResponseWriter) {
	a.clientError(w, http.StatusNotFound)
}

func lastDay(month time.Month) int {
	switch month {
	case time.April, time.June, time.September, time.November:
		return 30
	case time.February:
		year := time.Now().Year()
		if year%4 == 0 || year%100 == 0 || year%400 == 0 {
			return 29
		} else {
			return 28
		}
	default:
		return 31
	}
}

func getDate(dateStr string, numDay int) (date time.Time, err error) {
	if dateStr == "" {
		date = time.Date(time.Now().Year(), time.Now().Month(), numDay, 0, 0, 0, 0, time.Local)
		return
	} else {
	
	}
	date, err = time.Parse("2006-01-02", dateStr)
	if err != nil {
		return
	}
	return
}

// calcSectorsInGraph is used to calculate path's coordinates attributes for svg drawing
// returned data is slice of path's. single slice is a slice of attributes
func calcSectorsInGraph(total int, sectors []types.Statistics) [][]string {
	colors := []string{"red", "blue", "brown", "green", "black", "yellow", "pink", "white", "orange"}
	result := make([][]string, len(sectors))
	type point struct {
		x, y int
	}
	center := point{150, 150}
	radius := 110
	var degree float64
	xStart := 150
	yStart := 40
	if len(sectors) == 1 {
		return [][]string{{fmt.Sprintf("M%d,%d L%d,%d A%d,%d 0 1 1 %d,%d A%d,%d 0 1 1 %d,%d z",
			center.x, center.y, xStart, yStart, radius, radius, xStart, yStart+2*radius, radius, radius, xStart, yStart),
			"red", fmt.Sprintf("%d", int(sectors[0].SumSubcat)), sectors[0].Cat}}
	}
	for i := 0; i < len(sectors); i++ {
		rotFlag := 0
		sectorDegree := 2 * math.Pi * float64(sectors[i].SumSubcat) / float64(total)
		degree += sectorDegree
		if sectorDegree >= math.Pi {
			rotFlag = 1
		}
		x := center.x + int(float64(radius)*math.Sin(degree))
		y := center.y - int(float64(radius)*math.Cos(degree))
		path := fmt.Sprintf("M%d,%d L%d,%d A%d,%d 0 %d 1 %d,%d z",
			center.x, center.y, xStart, yStart, radius, radius, rotFlag, x, y)
		color := colors[i%len(colors)]
		result[i] = []string{path, color}
		xStart = x
		yStart = y
	}
	return result
}

func convertCheckToExpenseAddTypes(check *types.Check) (expenses []*types.ExpenseAdd) {
	expenses = make([]*types.ExpenseAdd, len(check.Items))
	for i := 0; i < len(expenses); i++ {
		expenses[i] = &types.ExpenseAdd{
			Date:   check.Date,
			Name:   check.Items[i].Name,
			Subcat: types.SubCategories[check.Items[i].Subcat],
			City:   check.City,
			Online: false,
			Count:  fmt.Sprintf("%f", check.Items[i].Quantity),
			Price:  fmt.Sprintf("%f", check.Items[i].Price),
			NDS:    check.Items[i].Nds,
		}
	}
	return
}
