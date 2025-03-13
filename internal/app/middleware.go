package app

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
)

// secureHeaders adds extra headers
func (a *App) secureHeaders(c *fiber.Ctx) error {
	c.Set("X-XSS-Protection", "1; mode=block")
	c.Set("X-Frame-Options", "deny")
	return c.Next()
}

func (a *App) logRequest(c *fiber.Ctx) error {
	a.logger.Printf("%s %s %s", c.IP(), c.Method(), c.OriginalURL())
	return c.Next()
}

func (a *App) recoverPanic(c *fiber.Ctx) error {
	defer func() {
		if err := recover(); err != nil {
			c.Set("Connection", "close")
			a.serverError(c, fmt.Errorf("%v", err))
		}
	}()
	return c.Next()
}

func (a *App) cacheMiddleware(c *fiber.Ctx) error {
	if len(a.cache.constDataCache["cities"]) == 0 {
		cities, err := a.db.GetCities()
		if err != nil {
			a.serverError(c, err)
			return nil
		}
		a.cache.constDataCache["cities"] = cities
	}
	if len(a.cache.constDataCache["names"]) == 0 {
		names, err := a.db.GetExpensesNames()
		if err != nil {
			a.serverError(c, err)
			return nil
		}
		a.cache.constDataCache["names"] = names
	}
	if len(a.cache.constDataCache["categories"]) == 0 {
		categories, err := a.db.GetCategories()
		if err != nil {
			a.serverError(c, err)
			return nil
		}
		a.cache.constDataCache["categories"] = categories
	}
	return c.Next()
}
