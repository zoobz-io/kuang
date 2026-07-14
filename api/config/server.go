package config

import "github.com/zoobz-io/check"

// Server holds the HTTP listener address and credential store location.
type Server struct {
	Host   string `env:"KUANG_HOST" default:"localhost"`
	DBPath string `env:"KUANG_DB_PATH" default:"data/kuang.db"`
	Port   int    `env:"KUANG_PORT" default:"8080"`
}

// Validate checks that all required fields are set and the port is in range.
func (c Server) Validate() error {
	return check.All(
		check.Str(c.Host, "host").Required().V(),
		check.Str(c.DBPath, "db_path").Required().V(),
		check.PortNumber(c.Port, "port"),
	).Err()
}
