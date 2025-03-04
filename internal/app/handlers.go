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
	
	if err := a.render(c.Writer, c.Request, "expenses.page.gohtml", &types.TemplateResult{
		DateLow:      dateLow.Format("2006-01-02"),
		DateHigh:     dateHigh.Format("2006-01-02"),
		ExpensesShow: expensesShow,
	}); err != nil {
		return
	}
}

func (a *App) ExpensesPost(c *gin.Context) {
	filter := types.FilterExpenses{}
	if err := c.Bind(&filter); err != nil {
		a.logger.Errorf("error while bind in ExpensesPost: %s", err.Error())
		a.clientError(c.Writer, http.StatusBadRequest)
		return
	}
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/expense?subcat=%s&date_low=%s&date_high=%s", filter.Category, filter.DateLow, filter.DateHigh))
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
		sumTotal += int(stats[i].SumCategory)
	}
	pngCategory := make([][4]string, len(stats))
	
	// sumSubcatsSlice is needed for proper calculations of sectors shape and showing this data in graph
	sumSubcatsSlice := make([]types.Statistics, 0, len(stats))
	for i := 0; i < len(stats); i++ {
		sumSubcatsSlice = append(sumSubcatsSlice, types.Statistics{
			Category:    stats[i].Category,
			SumCategory: stats[i].SumCategory,
		})
	}
	pngCategory = calcSectorsInGraph(sumTotal, sumSubcatsSlice)
	
	if err := a.render(c.Writer, c.Request, "statistics.page.gohtml", &types.TemplateResult{
		DateLow:  dateLow.Format("2006-01-02"),
		DateHigh: dateHigh.Format("2006-01-02"),
		Statistics: types.StatAndSum{
			Sum:         sumTotal,
			Statistics:  stats,
			PngCategory: pngCategory,
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
	cities, err := a.db.GetCities()
	if err != nil {
		a.serverError(c.Writer, err)
		return
	}
	
	names, err := a.db.GetExpensesNames()
	if err != nil {
		a.serverError(c.Writer, err)
		return
	}
	categories, err := a.db.GetCategories()
	if err != nil {
		a.serverError(c.Writer, err)
		return
	}
	if err := a.render(c.Writer, c.Request, "add.page.gohtml", &types.TemplateResult{AddShow: types.AddExpenseShowForm{
		Date:     time.Now().Format("2006-01-02"),
		Cities:   cities,
		Name:     names,
		Category: categories,
		Online:   []string{"true", "false"},
		Form:     &types.Form{},
	}}); err != nil {
		return
	}
}

func (a *App) AddExpensePost(c *gin.Context) {
	ea := types.ExpenseAdd{}
	if err := c.Bind(&ea); err != nil {
		a.clientError(c.Writer, http.StatusBadRequest)
		return
	}
	form := types.NewForm(&ea)
	form.Required("date", "name", "category", "city", "count", "price")
	
	cities, err := a.db.GetCities()
	if err != nil {
		a.serverError(c.Writer, err)
		return
	}
	names, err := a.db.GetExpensesNames()
	if err != nil {
		a.serverError(c.Writer, err)
		return
	}
	categories, err := a.db.GetCategories()
	if err != nil {
		a.serverError(c.Writer, err)
		return
	}
	
	if !form.Valid() {
		a.render(c.Writer, c.Request, "add.page.gohtml", &types.TemplateResult{AddShow: types.AddExpenseShowForm{
			Date:     time.Now().Format("2006-01-02"),
			Cities:   cities,
			Name:     names,
			Category: categories,
			Online:   []string{"true", "false"},
			Form:     form,
		}})
		return
	}
	
	if err = a.db.AddExpense(&ea); err != nil {
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
				subcats := strings.Split(mpd.Value["subcat"][0], "|")
				if len(check.Items) != len(subcats) {
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
					check.Items[i].Category = subcats[i]
				}
				check.City = mpd.Value["city"][0]
				expensesToAdd := convertCheckToExpenseAddTypes(&check)
				for i := 0; i < len(expensesToAdd); i++ {
					if err = a.db.AddExpense(expensesToAdd[i]); err != nil {
						a.serverError(c.Writer, err)
						return
					}
				}
			}
		}
	}
	
	c.Redirect(http.StatusSeeOther, "/add")
}

func (a *App) SearchGet(c *gin.Context) {
	
	if err := a.render(c.Writer, c.Request, "search.page.gohtml", &types.TemplateResult{
		SearchResult: nil,
	}); err != nil {
		return
	}
}

func (a *App) SearchPost(c *gin.Context) {
	var search string
	if err := c.Request.ParseForm(); err != nil {
		a.logger.Errorf("error while bind in Search: %s", err.Error())
		a.clientError(c.Writer, http.StatusBadRequest)
		return
	}
	search = c.Request.Form["search"][0]
	if search == "" {
		a.clientError(c.Writer, http.StatusBadRequest)
		return
	}
	searchResult, err := a.db.SearchExpense(search)
	if err != nil {
		if errors.Is(err, types.ErrNoRecord) {
			a.notFound(c.Writer)
		} else {
			a.serverError(c.Writer, err)
			return
		}
	}
	
	if err := a.render(c.Writer, c.Request, "search.page.gohtml", &types.TemplateResult{
		SearchResult: searchResult,
	}); err != nil {
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
	ts, ok := a.cache.templateCache[name]
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
