package app

import (
	"html/template"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
)

type cache struct {
	templateCache  map[string]*template.Template
	constDataCache map[string][]string
	mutex          *sync.Mutex
	eTag           atomic.Uint32
}

func newCache(templateDir string) (*cache, error) {
	templateCache, err := newTemplateCache(templateDir)
	if err != nil {
		return nil, err
	}
	constDataCache := make(map[string][]string)
	return &cache{
		templateCache:  templateCache,
		eTag:           atomic.Uint32{},
		constDataCache: constDataCache,
		mutex:          &sync.Mutex{},
	}, nil
}

func newTemplateCache(dir string) (map[string]*template.Template, error) {
	// Initialize a new map to act as the cache.
	cache := map[string]*template.Template{}
	// Use the filepath.Glob function to get a slice of all filepaths with
	// the extension '.page.gohtml'. This essentially gives us a slice of all the
	// 'page' templates for the application.
	pages, err := filepath.Glob(filepath.Join(dir, "*.page.gohtml"))
	if err != nil {
		return nil, err
	}
	// Loop through the pages one-by-one.
	for _, page := range pages {
		// Extract the file name (like 'home.page.gohtml') from the full file path
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
		ts, err = ts.ParseGlob(filepath.Join(dir, "*.layout.gohtml"))
		if err != nil {
			return nil, err
		}
		// Use the ParseGlob method to add any 'partial' templates to the
		// template set (in our case, it's just the 'footer' partial at the
		// moment).
		ts, err = ts.ParseGlob(filepath.Join(dir, "*.partial.gohtml"))
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

func (c *cache) updateEtag() {
	c.eTag.Add(1)
}

func (c *cache) getEtag() uint32 {
	return c.eTag.Load()
}

func (c *cache) eTagStr() string {
	return strconv.Itoa(int(c.getEtag()))
}
