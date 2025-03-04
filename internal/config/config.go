package config

type AppConfig struct {
	Addr      string // default 5000
	StaticDir string // default ./static
	LogFile   string // default stdout
	DBType    string // default sqlite
	DBPath    string // default ./db
}

func NewAppConfig(addr, staticDir, logfile, dbName string) *AppConfig {
	config := new(AppConfig)
	config.Addr = addr
	config.StaticDir = staticDir
	config.LogFile = logfile
	config.DBType = dbName
	return config
}
