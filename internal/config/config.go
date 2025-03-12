package config

type AppConfig struct {
	Addr        string // default 5000
	StaticDir   string // default ./static/
	TemplateDir string // defalut ./static/html/
	LogDir      string // default ./logs/
	LogFile     string // default stdout
	DBType      string // default sqlite
	DBPath      string // default ./db/
	DBName      string // default "expenses.db"
}

func NewAppConfig(addr, staticDir, templateDir, logDir, logFile, dbType, dbPath, dbName string) *AppConfig {
	return &AppConfig{
		Addr:        addr,
		StaticDir:   staticDir,
		TemplateDir: templateDir,
		LogDir:      logDir,
		LogFile:     logFile,
		DBType:      dbType,
		DBPath:      dbPath,
		DBName:      dbName,
	}
}
