package admin

// Config holds admin server settings loaded from the environment.
type Config struct {
	Host   string `env:"ADMIN_HOST" default:"localhost"`
	Port   int    `env:"ADMIN_PORT" default:"9090"`
	DBPath string `env:"ADMIN_DB_PATH" default:"data/kuang.db"`
}
