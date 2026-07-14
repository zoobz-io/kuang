package matrix

import "github.com/zoobz-io/check"

// Config holds Matrix homeserver settings.
type Config struct {
	Homeserver string `env:"KUANG_MATRIX_HOMESERVER"`
}

// Validate checks that required fields are set and the homeserver is well-formed.
func (a Config) Validate() error {
	return check.All(
		check.Str(a.Homeserver, "homeserver").Required().URL().V(),
	).Err()
}
