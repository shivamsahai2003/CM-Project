package config

import "os"

// Config holds all application configuration
type Config struct {
	DBDsn      string
	ServerAddr string
	APIBaseURL string
}

// Load returns the application configuration from environment variables
func Load() *Config {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "root:qwerty1@U@tcp(127.0.0.1:3306)/adservingproj?parseTime=true&charset=utf8mb4&loc=Local"
	}

	addr := os.Getenv("SERVER_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	apiBase := os.Getenv("KEYWORD_API_BASE")
	if apiBase == "" {
		apiBase = "http://g-usw1b-kwd-api-realapi.srv.media.net/kbb/keyword_api.php"
	}

	return &Config{
		DBDsn:      dsn,
		ServerAddr: addr,
		APIBaseURL: apiBase,
	}
}
