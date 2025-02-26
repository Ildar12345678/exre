package app

import (
	"expenses2/internal/log"
	"expenses2/internal/config"
	"net/http"
	"expenses2/internal/db"
	"github.com/gin-gonic/gin"
	"context"
)

type App struct {
	logger     *log.Logger
	conf       *config.AppConfig
	srv        *http.Server
	db         *dblayer.DB
	cache      *cache
	shutdownCh chan struct{}
}

func NewApp(addr, staticDir, logfile string) (*App, error) {
	conf := config.NewAppConfig(addr, staticDir, logfile)
	db, err := dblayer.NewDB()
	if err != nil {
		return nil, err
	}
	cache, err := newCache("./html", db)
	if err != nil {
		return nil, err
	}
	logger, err := log.NewLog(conf.LogFile)
	if err != nil {
		return nil, err
	}
	return &App{
		logger: logger,
		conf:   conf,
		db:     db,
		cache:  cache,
	}, nil
}

func (a *App) routes() http.Handler {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	
	router.Use(a.logRequest(), a.recoverPanic(), a.secureHeaders())
	router.GET("/expense", a.ExpensesGet)
	router.POST("/expense", a.ExpensesPost)
	router.GET("/stat", a.StatGet)
	router.POST("/stat", a.StatPost)
	router.GET("/add", a.AddExpenseGet)
	router.POST("/add", a.AddExpensePost)
	router.POST("/upload", a.UploadExpensesFromJson)
	router.GET("/search", a.SearchGet)
	router.POST("/search", a.SearchPost)
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
	if err := dblayer.CloseDB(a.db); err != nil {
		a.logger.Errorf("error while closing DB:", err.Error())
	}
	return a.srv.Shutdown(ctx)
}
