package main

import (
	"flag"
	"expenses2/internal/app"
	"os"
	"os/signal"
	"syscall"
	"log"
	"net/http"
	"expenses2/internal/config"
)

func main() {
	addr := flag.String("addr", ":5000", "listening app port")
	staticDir := flag.String("static", "./static", "location of static files")
	logFile := flag.String("log", "stdout", "file to log messages")
	db := flag.String("db", "sqlite", "choose db to store your data")
	
	flag.Parse()
	
	conf := config.NewAppConfig(*addr, *staticDir, *logFile, *db)
	
	newApp, err := app.NewApp(conf)
	if err != nil {
		panic(err)
	}
	
	// idleConnsClosed := make(chan struct{})
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("Shutting down server...")
		
		// We received an interrupt signal, shut down.
		if err := newApp.Shutdown(); err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("HTTP server Shutdown: %v", err)
		}
		// close(idleConnsClosed)
	}()
	
	if err := newApp.StartServer(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}
	log.Println("Server exiting")
	
	// <-idleConnsClosed
}
