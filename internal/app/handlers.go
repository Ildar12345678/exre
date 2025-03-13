package app

import (
	"bytes"
	"errors"
	"expenses2/internal/types"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {

}

func NewHandler() {
	
}

func (a *App) ExpensesGet(c *fiber.Ctx) error {
	var dateLow, dateHigh time.Time
	var err error
	dateLow, err = getDate(c.Query("date_low"), 1)
	if err != nil {
		a.clientError(c, http.StatusBadRequest)
		return nil
	}
	dateHigh, err = getDate(c.Query("date_high"), lastDay(time.Now().Month()))
	if err != nil {
		a.clientError(c, http.StatusBadRequest)
		return nil
	}

	if dateHigh.Sub(dateLow) < 0 {
		a.clientError(c, http.StatusBadRequest)
		return nil
	}

	if match := c.Get("If-None-Match"); match == a.cache.eTagStr() {
		c.Status(http.StatusNotModified) // Tell browser to use cache
		return nil
	}
	expensesShow, err := a.db.GetExpenses(dateLow, dateHigh)
	if err != nil {
		if errors.Is(err, types.ErrNoRecord) {
			a.notFound(c)
		} else {
			a.serverError(c, err)
		}
		return nil
	}

	c.Set("ETag", a.cache.eTagStr())

	return a.render(c, "expenses.page.gohtml", &types.TemplateResult{
		URL:          "/expense",
		DateLow:      dateLow.Format("2006-01-02"),
		DateHigh:     dateHigh.Format("2006-01-02"),
		ExpensesShow: expensesShow,
	})
}

func (a *App) ExpensesUpdate(c *fiber.Ctx) error {
	ea := types.ExpenseShow{}
	if err := c.BodyParser(&ea); err != nil {
		a.clientError(c, http.StatusBadRequest)
		return nil
	}
	if err := a.db.UpdateExpense(&ea); err != nil {
		a.serverError(c, err)
		return nil
	}
	if err := a.updateCache(); err != nil {
		a.serverError(c, err)
		return nil
	}
	a.cache.updateEtag()
	return c.Redirect("/expense", http.StatusSeeOther)
}

func (a *App) ExpensesDelete(c *fiber.Ctx) error {
	ea := types.ExpenseShow{}
	if err := c.BodyParser(&ea); err != nil {
		a.clientError(c, http.StatusBadRequest)
		return nil
	}
	if err := a.db.DeleteExpense(ea.ID); err != nil {
		a.serverError(c, err)
		return nil
	}
	if err := a.updateCache(); err != nil {
		a.serverError(c, err)
		return nil
	}
	a.cache.updateEtag()
	return c.Redirect("/expense", http.StatusSeeOther)
}


func (a *App) StatGet(c *fiber.Ctx) error {
	var dateLow, dateHigh time.Time
	var err error

	dateLow, err = getDate(c.Query("date_low"), 1)
	if err != nil {
		a.clientError(c, http.StatusBadRequest)
		return nil
	}
	dateHigh, err = getDate(c.Query("date_high"), lastDay(time.Now().Month()))
	if err != nil {
		a.clientError(c, http.StatusBadRequest)
		return nil
	}

	if dateHigh.Sub(dateLow) < 0 {
		a.clientError(c, http.StatusBadRequest)
		return nil
	}
	
	if match := c.Get("If-None-Match"); match == a.cache.eTagStr() {
		c.Status(http.StatusNotModified) // Tell browser to use cache
		return nil
	}

	stats, err := a.db.GetStatistics(dateLow, dateHigh)
	if err != nil {
		if errors.Is(err, types.ErrNoRecord) {
			a.notFound(c)
		} else {
			a.serverError(c, err)
		}
		return nil
	}

	// let's calc total sum of expenses in specified period
	var sumTotal int
	for i := 0; i < len(stats); i++ {
		sumTotal += int(stats[i].SumCategory)
	}

	// sumSubcatsSlice is needed for proper calculations of sectors shape and showing this data in graph
	sumSubcatsSlice := make([]types.Statistics, 0, len(stats))
	for i := 0; i < len(stats); i++ {
		sumSubcatsSlice = append(sumSubcatsSlice, types.Statistics{
			Category:    stats[i].Category,
			SumCategory: stats[i].SumCategory,
		})
	}

	c.Set("ETag", a.cache.eTagStr())

	return a.render(c, "statistics.page.gohtml", &types.TemplateResult{
		URL:      "/expense/stat",
		DateLow:  dateLow.Format("2006-01-02"),
		DateHigh: dateHigh.Format("2006-01-02"),
		Statistics: types.StatAndSum{
			Sum:         sumTotal,
			Statistics:  stats,
			PngCategory: calcSectorsInGraph(sumTotal, sumSubcatsSlice),
		}})
}

func (a *App) Dates(c *fiber.Ctx) error {
	filter := types.FilterExpenses{}
	if err := c.BodyParser(&filter); err != nil {
		a.clientError(c, http.StatusBadRequest)
		return nil
	}
	return c.Redirect(fmt.Sprintf("%s?date_low=%s&date_high=%s&subcat=%s", filter.URL, filter.DateLow, filter.DateHigh, filter.Category),
		http.StatusSeeOther)
}

func (a *App) AddExpenseGet(c *fiber.Ctx) error {
	a.cache.mutex.Lock()
	cities := a.cache.constDataCache["cities"]
	names := a.cache.constDataCache["names"]
	categories := a.cache.constDataCache["categories"]
	a.cache.mutex.Unlock()
	return a.render(c, "add.page.gohtml", &types.TemplateResult{AddShow: types.AddExpenseShowForm{
		Date:     time.Now().Format("2006-01-02"),
		Cities:   cities,
		Name:     names,
		Category: categories,
		Online:   []string{"true", "false"},
		Form:     &types.FormAddExpense{},
	}})
}

func (a *App) AddExpensePost(c *fiber.Ctx) error {
	ea := types.ExpenseAdd{}
	if err := c.BodyParser(&ea); err != nil {
		a.clientError(c, http.StatusBadRequest)
		return nil
	}
	form := types.NewForm(&ea)
	form.Required("date", "name", "category", "city", "count", "price")

	a.cache.mutex.Lock()
	cities := a.cache.constDataCache["cities"]
	names := a.cache.constDataCache["names"]
	categories := a.cache.constDataCache["categories"]
	a.cache.mutex.Unlock()

	if !form.Valid() {
		return a.render(c, "add.page.gohtml", &types.TemplateResult{AddShow: types.AddExpenseShowForm{
			Date:     time.Now().Format("2006-01-02"),
			Cities:   cities,
			Name:     names,
			Category: categories,
			Online:   []string{"true", "false"},
			Form:     form,
		}})
	}

	if err := a.db.AddExpense(&ea); err != nil {
		a.serverError(c, err)
		return nil
	}

	// Update cache after adding new expense
	if err := a.updateCache(); err != nil {
		a.serverError(c, err)
		return nil
	}

	a.cache.updateEtag()
	return c.Redirect("/expense/add", http.StatusSeeOther)
}

func (a *App) UploadExpensesFromJson(c *fiber.Ctx) error {
	ea := &types.ExpenseAddFromCheck{}
	form := types.NewForm(ea)
	// todo save file in session
	mpd, err := c.MultipartForm()
	if err != nil {
		form.Errors.Add("file", "incorrect form parse")
		return a.render(c, "add.page.gohtml", &types.TemplateResult{AddShow: types.AddExpenseShowForm{
			Date: time.Now().Format("2006-01-02"),
			Form: form,
		}})
	}
	files := mpd.File["file"]
	if len(files) == 0 {
		form.Errors.Add("file", "no files uploaded")
		return a.render(c, "add.page.gohtml", &types.TemplateResult{AddShow: types.AddExpenseShowForm{
			Date: time.Now().Format("2006-01-02"),
			Form: form,
		}})
	}

	ea.File = files[0]
	form.Multipart = ea.File
	check, err := form.CheckFile()
	if err != nil {
		a.serverError(c, err)
	}

	categories := mpd.Value["category_check"]
	if len(categories) == 0 {
		form.Errors.Add("category_check", "empty input is not permitted")
		return a.render(c, "add.page.gohtml", &types.TemplateResult{AddShow: types.AddExpenseShowForm{
			Date: time.Now().Format("2006-01-02"),
			Form: form,
		}})
	}
	form.Values.Add("category_check", categories[0])
	form.Required("category_check")

	categorySlice := strings.Split(categories[0], ",")
	if len(check.Items) != len(categorySlice) {
		form.Errors.Add("category_check", "enter correct amount of categories")
		return a.render(c, "add.page.gohtml", &types.TemplateResult{AddShow: types.AddExpenseShowForm{
			Date: time.Now().Format("2006-01-02"),
			Form: form,
		}})
	}
	for i := range check.Items {
		check.Items[i].Price /= 100
		if check.Items[i].Nds == 1 {
			check.Items[i].Nds = 20
		} else {
			check.Items[i].Nds = 10
		}
		check.Items[i].Category = categorySlice[i]
	}
	cities := mpd.Value["city_check"]
	if len(cities) == 0 {
		form.Errors.Add("city_check", "enter correct city name")
		return a.render(c, "add.page.gohtml", &types.TemplateResult{AddShow: types.AddExpenseShowForm{
			Date: time.Now().Format("2006-01-02"),
			Form: form,
		}})
	}

	if !form.Valid() {
		return a.render(c, "add.page.gohtml", &types.TemplateResult{AddShow: types.AddExpenseShowForm{
			Cities:   cities,
			Category: categories,
			Form:     form,
		}})
	}

	check.City = cities[0]
	expensesToAdd := convertCheckToExpenseAddTypes(check)
	for i := 0; i < len(expensesToAdd); i++ {
		if err = a.db.AddExpense(expensesToAdd[i]); err != nil {
			a.serverError(c, err)
			return nil
		}
	}

	// Update cache after uploading expenses
	if err := a.updateCache(); err != nil {
		a.serverError(c, err)
		return nil
	}

	a.cache.updateEtag()
	return c.Redirect("/expense/add", http.StatusSeeOther)
}

func (a *App) SearchGet(c *fiber.Ctx) error {
	return a.render(c, "search.page.gohtml", &types.TemplateResult{
		SearchResult: nil,
	})
}

func (a *App) SearchPost(c *fiber.Ctx) error {
	search := c.FormValue("search")
	if search == "" {
		a.clientError(c, http.StatusBadRequest)
		return nil
	}
	searchResult, err := a.db.SearchExpense(search)
	if err != nil {
		if errors.Is(err, types.ErrNoRecord) {
			a.notFound(c)
		} else {
			a.serverError(c, err)
			return nil
		}
	}

	return a.render(c, "search.page.gohtml", &types.TemplateResult{
		SearchResult: searchResult,
	})
}

// addDefaultData is used to add request.URL.Path to data which is transfer to template (for dateInput template)
// func (a *App) addDefaultData(tr *types.TemplateResult, r *http.Request) *types.TemplateResult {
// if tr == nil {
// tr = &types.TemplateResult{}
// }
// tr.URL = r.URL.Path
// return tr
// }

// render is general function to send data to responseWriter
func (a *App) render(c *fiber.Ctx, name string, tr *types.TemplateResult) error {
	// Retrieve the appropriate template set from the cache based on the page name
	// (like 'home.page.gohtml'). If no entry exists in the cache with the
	// provided name, call the serverError helper method that we made earlier.
	ts, ok := a.cache.templateCache[name]
	if !ok {
		a.serverError(c, fmt.Errorf("the template %s does not exist", name))
		return fmt.Errorf("no required template")
	}

	// Initialize a new buffer.
	buf := new(bytes.Buffer)

	// firstly execute to the buffer. if error occurs no data is sent to client
	err := ts.Execute(buf, tr)
	if err != nil {
		a.serverError(c, err)
		return err
	}

	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.Status(http.StatusOK).Send(buf.Bytes())
}

func (a *App) updateCache() error {
	cities, err := a.db.GetCities()
	if err != nil {
		return err
	}
	names, err := a.db.GetExpensesNames()
	if err != nil {
		return err
	}
	categories, err := a.db.GetCategories()
	if err != nil {
		return err
	}

	a.cache.mutex.Lock()
	a.cache.constDataCache["cities"] = cities
	a.cache.constDataCache["names"] = names
	a.cache.constDataCache["categories"] = categories
	a.cache.mutex.Unlock()

	return nil
}
