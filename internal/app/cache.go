package app

import (
	"html/template"
	"path/filepath"
	dblayer "expenses2/internal/db"
	"fmt"
	"errors"
)

type cache struct {
	templateCache map[string]*template.Template
	dbCache       map[string][]string
	db            *dblayer.DB
}

func newCache(templateDir string, db *dblayer.DB) (*cache, error) {
	templateCache, err := newTemplateCache(templateDir)
	if err != nil {
		return nil, err
	}
	dbCache := map[string][]string{
		"city":    make([]string, 0, 10),
		"subcat":  make([]string, 0, 40),
		"expense": make([]string, 0, 1000),
	}
	
	cities, err := db.GetCity()
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(cities); i++ {
		dbCache["city"] = append(dbCache["city"], cities[i].City)
	}
	subcats, err := db.GetSubcat()
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(subcats); i++ {
		dbCache["subcat"] = append(dbCache["subcat"], fmt.Sprintf("%d-%s", subcats[i].ID, subcats[i].Name))
	}
	expenseNames, err := db.GetExpensesNames()
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(expenseNames); i++ {
		dbCache["expense"] = append(dbCache["expense"], expenseNames[i].Name)
	}
	return &cache{
		templateCache: templateCache,
		dbCache:       dbCache,
		db:            db,
	}, nil
}

func newTemplateCache(dir string) (map[string]*template.Template, error) {
	// Initialize a new map to act as the cache.
	cache := map[string]*template.Template{}
	// Use the filepath.Glob function to get a slice of all filepaths with
	// the extension '.page.tmpl'. This essentially gives us a slice of all the
	// 'page' templates for the application.
	pages, err := filepath.Glob(filepath.Join(dir, "*.page.tmpl"))
	if err != nil {
		return nil, err
	}
	// Loop through the pages one-by-one.
	for _, page := range pages {
		// Extract the file name (like 'home.page.tmpl') from the full file path
		// and assign it to the name variable.
		name := filepath.Base(page)
		// Parse the page template file in to a template set.
		ts, err := template.ParseFiles(page)
		if err != nil {
			return nil, err
		}
		// Use the ParseGlob method to add any 'layout' templates to the
		// template set (in our case, it's just the 'base' layout at the
		// moment).
		ts, err = ts.ParseGlob(filepath.Join(dir, "*.layout.tmpl"))
		if err != nil {
			return nil, err
		}
		// Use the ParseGlob method to add any 'partial' templates to the
		// template set (in our case, it's just the 'footer' partial at the
		// moment).
		ts, err = ts.ParseGlob(filepath.Join(dir, "*.partial.tmpl"))
		if err != nil {
			return nil, err
		}
		// Add the template set to the cache, using the name of the page
		// (like 'home.page.tmpl') as the key.
		cache[name] = ts
	}
	// Return the map.
	return cache, nil
}

func (c *cache) updateCache(keyToUpdate string, ids ...any) (bool, error) {
	switch keyToUpdate {
	case "city":
		cities, err := c.db.GetCity()
		if err != nil {
			return false, err
		}
		c.dbCache["city"] = nil
		for i := 0; i < len(cities); i++ {
			c.dbCache["city"] = append(c.dbCache["city"], cities[i].City)
		}
		return true, nil
	case "subcat":
		subcats, err := c.db.GetSubcat()
		if err != nil {
			return false, err
		}
		c.dbCache["subcat"] = nil
		for i := 0; i < len(subcats); i++ {
			c.dbCache["subcat"] = append(c.dbCache["subcat"], fmt.Sprintf("%d-%s", subcats[i].ID, subcats[i].Name))
		}
		return true, nil
	case "expense":
		expenseNames, err := c.db.GetExpensesNames(ids...)
		if err != nil {
			return false, err
		}
		for i := 0; i < len(expenseNames); i++ {
			c.dbCache["expense"] = append(c.dbCache["expense"], expenseNames[i].Name)
		}
		return true, nil
	default:
		return false, errors.New("incorrect key")
	}
}
