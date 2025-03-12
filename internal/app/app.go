package app

import (
	"expenses2/internal/config"
	"expenses2/internal/db"
	"expenses2/internal/log"

	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

type App struct {
	logger  *log.Logger
	conf    *config.AppConfig
	handler *Handler
	srv     *fiber.App
	db      db.DB
	cache   *cache
}

func NewApp(appConfig *config.AppConfig) (*App, error) {
	db, err := db.NewDB(appConfig.DBType, appConfig.DBPath, appConfig.DBName)
	if err != nil {
		return nil, err
	}
	err = db.Initialize()
	if err != nil {
		return nil, err
	}
	cache, err := newCache(appConfig.TemplateDir)
	if err != nil {
		return nil, err
	}
	logger, err := log.NewLog(appConfig.LogDir, appConfig.LogFile)
	if err != nil {
		return nil, err
	}
	return &App{
		logger: logger,
		conf:   appConfig,
		db:     db,
		cache:  cache,
	}, nil
}

func (a *App) routes() *fiber.App {
	app := fiber.New()
	
	// Middleware
	app.Use(a.logRequest, a.recoverPanic, a.secureHeaders)
	
	// Routes
	app.Get("/expense", a.ExpensesGet)
	app.Get("/expense/stat", a.StatGet)
	app.Post("/expense/dates", a.Dates)
	app.Get("/expense/add", a.AddExpenseGet)
	app.Post("/expense/add", a.AddExpensePost)
	app.Post("/expense/upload", a.UploadExpensesFromJson)
	app.Get("/expense/search", a.SearchGet)
	app.Post("/expense/search", a.SearchPost)
	app.Static("/expense/static", a.conf.StaticDir)
	
	return app
}

func (a *App) StartServer() error {
	a.logger.Printf("Starting server on %s\n", a.conf.Addr)
	app := a.routes()
	a.srv = app
	return app.Listen(a.conf.Addr)
}

func (a *App) Shutdown() error {
	if err := a.db.Close(); err != nil {
		a.logger.Errorf("error while closing DB:", err.Error())
	}
	return a.srv.Shutdown()
}
