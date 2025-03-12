package main

import (
	"expenses2/internal/app"
	"expenses2/internal/config"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	addr        = flag.String("addr", ":5000", "listening app port")
	staticDir   = flag.String("static", "./static/", "location of static files")
	templateDir = flag.String("template", "./static/html/", "location of templates")
	logDir      = flag.String("logdir", "./logs/", "logs directory")
	logFile     = flag.String("log", "stdout", "file to log messages")
	dbType      = flag.String("dbtype", "sqlite", "choose db type to store your data")
	dbPath      = flag.String("dbpath", "./db/", "choose db path to initialize and store your data")
	dbName      = flag.String("dbname", "expenses.db", "choose db name")
)

func main() {
	flag.Parse()

	conf := config.NewAppConfig(*addr, *staticDir, *templateDir, *logDir, *logFile, *dbType, *dbPath, *dbName)
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
