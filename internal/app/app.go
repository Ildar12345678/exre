package app

import (
	"expenses2/internal/log"
	"expenses2/internal/config"
	"net/http"
	"expenses2/internal/db"
	"html/template"
	"github.com/gin-gonic/gin"
	"context"
)

type App struct {
	logger        *log.Logger
	conf          *config.AppConfig
	srv           *http.Server
	db            *dblayer.DB
	templateCache map[string]*template.Template
	shutdownCh    chan struct{}
}

func NewApp(addr, staticDir, logfile string) (*App, error) {
	conf := config.NewAppConfig(addr, staticDir, logfile)
	db, err := dblayer.NewDB()
	if err != nil {
		return nil, err
	}
	tc, err := newTemplateCache("./html")
	if err != nil {
		return nil, err
	}
	logger, err := log.NewLog(conf.LogFile)
	if err != nil {
		return nil, err
	}
	return &App{
		logger:        logger,
		conf:          conf,
		db:            db,
		templateCache: tc,
	}, nil
}

func (a *App) routes() http.Handler {
	router := gin.New()
	
	router.Use(a.logRequest2(), a.recoverPanic2(), a.secureHeaders2())
	router.GET("/expense", a.ExpensesGet)
	router.POST("/expense", a.ExpensesPost)
	router.GET("/stat", a.StatGet)
	router.POST("/stat", a.StatPost)
	router.GET("/add", a.AddExpenseGet)
	router.POST("/add", a.AddExpensePost)
	router.POST("/upload", a.Upload)
	router.GET("/search", a.Search)
	router.GET("/click", a.Click)
	router.Static("/static/", a.conf.StaticDir)
	
	return router
}

func (a *App) StartServer() error {
	a.logger.Printf("Starting server on %s\n", a.conf.Addr)
	
	srv := &http.Server{
		Addr: a.conf.Addr,
		// ErrorLog: a.logger.ErrLogger,
		Handler: a.routes(),
	}
	a.srv = srv
	return srv.ListenAndServe()
}

func (a *App) Shutdown(ctx context.Context) error {
	return a.srv.Shutdown(ctx)
}

func RunAPIWithHandler() error {
	// r.GET("/city", handler.GetCity)
	// r.GET("/cat", handler.GetCat)
	// r.GET("/subcat", handler.GetSubcat)
	// r.GET("/supplier", handler.GetSupplier)
	// r.GET("/expname", handler.GetExpensesNames)
	// r.GET("/expenses", handler.GetExpenses)
	// r.GET("/stats", handler.GetStatistics)
	// r.POST("/stuff", handler.GetStuff)
	// r.POST("/date", handler.AddDate)
	// r.POST("/addexpense", handler.AddExpense)
	return nil
}
