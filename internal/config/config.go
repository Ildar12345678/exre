package config

type AppConfig struct {
	Addr      string `json:"addr"`       // default 5000
	StaticDir string `json:"static_dir"` // default ./static
	LogFile   string `json:"log_file"`   // default stdout
}

func NewAppConfig(addr, staticDir, logfile string) *AppConfig {
	config := new(AppConfig)
	config.Addr = addr
	config.StaticDir = staticDir
	config.LogFile = logfile
	return config
}
