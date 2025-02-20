package main

import (
	"flag"
	"expenses2/internal/app"
	"os"
	"os/signal"
	"syscall"
	"log"
	"time"
	"context"
	"net/http"
)

func main() {
	addr := flag.String("addr", ":5000", "listening app port")
	staticDir := flag.String("static", "./static", "location of static files")
	logFile := flag.String("log", "stdout", "file to log messages")
	
	flag.Parse()
	
	newApp, err := app.NewApp(*addr, *staticDir, *logFile)
	if err != nil {
		panic(err)
	}
	
	idleConnsClosed := make(chan struct{})
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("Shutting down server...")
		
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		// We received an interrupt signal, shut down.
		if err := newApp.Shutdown(ctx); err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()
	
	log.Println("Server exiting")
	
	if err := newApp.StartServer(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}
	<-idleConnsClosed
}
