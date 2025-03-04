package app

import (
	"fmt"
	"time"
	"github.com/gofiber/fiber/v2"
)

// secureHeaders adds extra headers
func (a *App) secureHeaders(c *fiber.Ctx) error {
	c.Set("X-XSS-Protection", "1; mode=block")
	c.Set("X-Frame-Options", "deny")
	c.Set("Cache-Control", "private, max-age=3600")
	c.Set("Last-Modified", time.Now().Format(time.RFC1123))
	return c.Next()
}

func (a *App) logRequest(c *fiber.Ctx) error {
	a.logger.Printf("%s - %s %s %s", c.IP(), c.Protocol(), c.Method(), c.OriginalURL())
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
