package admin

import "errors"

// Config holds admin server settings loaded from the environment.
type Config struct {
	Host   string `env:"ADMIN_HOST" default:"localhost"`
	Port   int    `env:"ADMIN_PORT" default:"9090"`
	DBPath string `env:"ADMIN_DB_PATH" default:"data/kuang.db"`
	APIKey string `env:"ADMIN_API_KEY"`
}

// Validate checks that required fields are set.
func (c Config) Validate() error {
	if c.APIKey == "" {
		return errors.New("admin: ADMIN_API_KEY is required")
	}
	return nil
}
