package app

import (
	"net/http"
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
)

// secureHeaders adds extra headers
func (a *App) secureHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("X-Frame-Options", "deny")
		c.Header("Cache-Control", "private, max-age=3600")
		c.Header("Last-Modified", time.Now().Format(http.TimeFormat))
	}
}

func (a *App) logRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		a.logger.Printf("%s - %s %s %s",
			c.ClientIP(), c.Request.Proto, c.Request.Method, c.Request.URL)
	}
}

func (a *App) recoverPanic() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Set a "Connection: close" header on the response.
				c.Header("Connection", "close")
				// Call the app.serverError helper method to return a 500
				// Internal Server response.
				a.serverError(c.Writer, fmt.Errorf("%s", err))
			}
		}()
	}
	// return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	defer func() {
	// 		// Use the builtin recover function to check if there has been a
	// 		// panic or not. If there has...
	// 		if err := recover(); err != nil {
	// 			// Set a "Connection: close" header on the response.
	// 			w.Header().Set("Connection", "close")
	// 			// Call the app.serverError helper method to return a 500
	// 			// Internal Server response.
	// 			a.serverError(w, fmt.Errorf("%s", err))
	// 		}
	// 	}()
	// 	next.ServeHTTP(w, r)
	// })
}
