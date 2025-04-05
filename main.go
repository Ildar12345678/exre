package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"expenses2/internal/app"
	"expenses2/internal/config"
)

var (
	addr        = flag.String("addr", ":5000", "listening app port")
	staticDir   = flag.String("static", "./static/", "location of static files")
	templateDir = flag.String("template", "./static/html/", "location of templates")
	logDir      = flag.String("logdir", "./logs/", "logs directory")
	logFile     = flag.String("log", "stdout", "file to log messages")
	dbPath      = flag.String("dbpath", "./db/", "choose db path to initialize and store your data")
	dbName      = flag.String("dbname", "expenses.db", "choose db name")
	defCity     = flag.String("defcity", "", "choose default city")
)

func main() {
	flag.Parse()

	conf := config.NewAppConfig(*addr, *staticDir, *templateDir, *logDir, *logFile, *dbPath, *dbName,*defCity)
	newApp, err := app.NewApp(conf)
	if err != nil {
		panic(err)
	}

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("Shutting down server...")

		if err := newApp.Shutdown(); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}
	}()

	if err := newApp.StartServer(); err != nil {
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}

	log.Println("Server exiting")
}
