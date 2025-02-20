package app

import (
	"net/http"
	"fmt"
	"github.com/gin-gonic/gin"
)

func (a *App) secureHeaders1(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Frame-Options", "deny")
		next.ServeHTTP(w, r)
	})
}

func (a *App) secureHeaders2() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("X-Frame-Options", "deny")
	}
}

func (a *App) logRequest1(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.logger.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())
		next.ServeHTTP(w, r)
	})
}

func (a *App) logRequest2() gin.HandlerFunc {
	// return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	a.logger.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())
	// 	next.ServeHTTP(w, r)
	// })
	return func(c *gin.Context) {
		a.logger.Printf("%s - %s %s %s", c.ClientIP(), c.Request.Proto, c.Request.Method, c.Request.RequestURI)
	}
}

func (a *App) recoverPanic1(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			// Use the builtin recover function to check if there has been a
			// panic or not. If there has...
			if err := recover(); err != nil {
				// Set a "Connection: close" header on the response.
				w.Header().Set("Connection", "close")
				// Call the app.serverError helper method to return a 500
				// Internal Server response.
				a.serverError(w, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (a *App) recoverPanic2() gin.HandlerFunc {
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
