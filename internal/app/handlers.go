package app

import (
	"net/http"
	"time"
	"errors"
	"expenses2/internal/types"
	"fmt"
	"bytes"
	"github.com/gin-gonic/gin"
	"encoding/json"
	"strings"
	"strconv"
)

func (a *App) ExpensesGet(c *gin.Context) {
	var dateLow, dateHigh time.Time
	var err error
	dateLow, err = getDate(c.Query("date_low"), 1)
	if err != nil {
		a.clientError(c.Writer, http.StatusBadRequest)
		return
	}
	dateHigh, err = getDate(c.Query("date_high"), lastDay(time.Now().Month()))
	if err != nil {
		a.clientError(c.Writer, http.StatusBadRequest)
		return
	}
	// subcat := c.Query("date_high")
	
	if dateHigh.Sub(dateLow) < 0 {
		a.clientError(c.Writer, http.StatusBadRequest)
		return
	}
	
	expensesShow, err := a.db.GetExpenses(dateLow, dateHigh)
	if err != nil {
		if errors.Is(err, types.ErrNoRecord) {
			a.notFound(c.Writer)
		} else {
			a.serverError(c.Writer, err)
		}
		return
	}
	
	if err := a.render(c.Writer, c.Request, "expenses.page.tmpl", &types.TemplateResult{ExpensesShow: expensesShow}); err != nil {
		return
	}
}

func (a *App) ExpensesPost(c *gin.Context) {
	filter := types.FilterExpenses{}
	if err := c.Bind(&filter); err != nil {
		fmt.Println(err)
		a.clientError(c.Writer, http.StatusBadRequest)
		return
	}
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/expense?subcat=%s&date_low=%s&date_high=%s", filter.Subcat, filter.DateLow, filter.DateHigh))
}

func (a *App) StatGet(c *gin.Context) {
	
	var dateLow, dateHigh time.Time
	var err error
	
	dateLow, err = getDate(c.Query("date_low"), 1)
	if err != nil {
		a.clientError(c.Writer, http.StatusBadRequest)
		return
	}
	dateHigh, err = getDate(c.Query("date_high"), lastDay(time.Now().Month()))
	if err != nil {
		a.clientError(c.Writer, http.StatusBadRequest)
		return
	}
	
	if dateHigh.Sub(dateLow) < 0 {
		a.clientError(c.Writer, http.StatusBadRequest)
		return
	}
	
	stats, err := a.db.GetStatistics(dateLow, dateHigh)
	if err != nil {
		if errors.Is(err, types.ErrNoRecord) {
			a.notFound(c.Writer)
		} else {
			a.serverError(c.Writer, err)
		}
		return
	}
	
	// let's calc total sum of expenses in specified period
	var sumTotal int
	for i := 0; i < len(stats); i++ {
		sumTotal += int(stats[i].SumSubcat)
	}
	// let's calc sum of expenses in all categories in specified period
	sumCats := make(map[string]int, len(stats))
	for i := 0; i < len(stats); i++ {
		sumCats[stats[i].Cat] += int(stats[i].SumSubcat)
	}
	pngSubcat := make([][]string, len(stats))
	pngCat := make([][]string, len(sumCats))
	
	// sumSubcatsSlice and sumCatsSlice are needed for proper calculations of sectors shape and showing this data in graph
	sumSubcatsSlice := make([]types.Statistics, 0, len(stats))
	for i := 0; i < len(stats); i++ {
		sumSubcatsSlice = append(sumSubcatsSlice, types.Statistics{
			Cat:       stats[i].Subcat,
			SumSubcat: stats[i].SumSubcat,
		})
	}
	sumCatsSlice := make([]types.Statistics, 0, len(sumCats))
	for k, v := range sumCats {
		sumCatsSlice = append(sumCatsSlice, types.Statistics{
			Cat:       k,
			SumSubcat: float32(v),
		})
	}
	pngSubcat = calcSectorsInGraph(sumTotal, sumSubcatsSlice)
	pngCat = calcSectorsInGraph(sumTotal, sumCatsSlice)
	
	if err := a.render(c.Writer, c.Request, "statistics.page.tmpl", &types.TemplateResult{Statistics: types.StatAndSum{
		Sum:        sumTotal,
		Statistics: stats,
		PngSubcat:  pngSubcat,
		PngCat:     pngCat,
	}}); err != nil {
		return
	}
}

func (a *App) StatPost(c *gin.Context) {
	dates := types.FilterExpenses{}
	if err := c.Bind(&dates); err != nil {
		a.clientError(c.Writer, http.StatusBadRequest)
		return
	}
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/stat?date_low=%s&date_high=%s", dates.DateLow, dates.DateHigh))
}

func (a *App) AddExpenseGet(c *gin.Context) {
	cities, err := a.db.GetCity()
	if err != nil {
		if errors.Is(err, types.ErrNoRecord) {
			a.notFound(c.Writer)
		} else {
			a.serverError(c.Writer, err)
		}
		return
	}
	subcats, err := a.db.GetSubcat()
	if err != nil {
		if errors.Is(err, types.ErrNoRecord) {
			a.notFound(c.Writer)
		} else {
			a.serverError(c.Writer, err)
		}
		return
	}
	expenseNames, err := a.db.GetExpensesNames()
	if err != nil {
		if errors.Is(err, types.ErrNoRecord) {
			a.notFound(c.Writer)
		} else {
			a.serverError(c.Writer, err)
		}
		return
	}
	form := types.AddShowForm{
		Date:        time.Now().Format("2006-01-02"),
		Cities:      cities,
		ExpenseName: expenseNames,
		Subcat:      subcats,
		Online:      []string{"true", "false"},
		Nds:         []string{"10", "20", "0"},
	}
	if err := a.render(c.Writer, c.Request, "add.page.tmpl", &types.TemplateResult{AddShow: form}); err != nil {
		return
	}
}

func (a *App) AddExpensePost(c *gin.Context) {
	ea := types.ExpenseAdd{}
	if err := c.Bind(&ea); err != nil {
		a.clientError(c.Writer, http.StatusBadRequest)
		return
	}
	
	if err := a.db.AddExpense(&ea); err != nil {
		a.serverError(c.Writer, err)
		return
	}
	c.Redirect(http.StatusSeeOther, "/add")
}

func (a *App) UploadExpensesFromJson(c *gin.Context) {
	mpd, err := c.MultipartForm()
	if err == nil {
		file, err := mpd.File["file"][0].Open()
		defer file.Close()
		if err == nil {
			ch := make([]types.MainDoc, 0)
			decoder := json.NewDecoder(file)
			err = decoder.Decode(&ch)
			if err == nil {
				check := ch[0].Ticket.Document.Receipt
				check.Date = strings.Split(check.Date, "T")[0]
				subcats := strings.Split(mpd.Value["subcat"][0], " ")
				if len(check.Items) != len(subcats) {
					fmt.Println(len(check.Items), len(subcats))
					a.clientError(c.Writer, http.StatusBadRequest)
					return
				}
				for i := range check.Items {
					check.Items[i].Price /= 100
					if check.Items[i].Nds == 1 {
						check.Items[i].Nds = 20
					} else {
						check.Items[i].Nds = 10
					}
					subcat, err := strconv.ParseInt(subcats[i], 10, 64)
					if err != nil {
						fmt.Println(err)
						a.clientError(c.Writer, http.StatusBadRequest)
						return
					}
					check.Items[i].Subcat = uint8(subcat)
				}
				check.City = mpd.Value["city"][0]
				expensesToAdd := convertCheckToExpenseAddTypes(&check)
				for i := 0; i < len(expensesToAdd); i++ {
					if err := a.db.AddExpense(expensesToAdd[i]); err != nil {
						a.serverError(c.Writer, err)
						return
					}
				}
			}
		}
	}
	c.Redirect(http.StatusSeeOther, "/add")
}

func (a *App) Search(c *gin.Context) {
	
	if err := a.render(c.Writer, c.Request, "search.page.tmpl", nil); err != nil {
		return
	}
}

// addDefaultData is used to add request.URL.Path to data which is transfer to template (for dateInput template)
func (a *App) addDefaultData(tr *types.TemplateResult, r *http.Request) *types.TemplateResult {
	if tr == nil {
		tr = &types.TemplateResult{}
	}
	tr.URL = r.URL.Path
	return tr
}

// render is general function to send data to responseWriter
func (a *App) render(w http.ResponseWriter, r *http.Request, name string, tr *types.TemplateResult) error {
	// Retrieve the appropriate template set from the cache based on the page name
	// (like 'home.page.tmpl'). If no entry exists in the cache with the
	// provided name, call the serverError helper method that we made earlier.
	ts, ok := a.templateCache[name]
	if !ok {
		a.serverError(w, fmt.Errorf("The template %s does not exist", name))
		return fmt.Errorf("no required template")
	}
	
	// Initialize a new buffer.
	buf := new(bytes.Buffer)
	
	// firstly execute to the buffer. if error occurs no data is sent to client
	err := ts.Execute(buf, a.addDefaultData(tr, r))
	if err != nil {
		a.serverError(w, err)
		return err
	}
	
	_, err = buf.WriteTo(w)
	return err
}
