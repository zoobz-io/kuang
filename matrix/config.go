// Package matrix provides a kuang module for Matrix messaging operations.
package matrix

import "errors"

// Config holds Matrix homeserver settings.
type Config struct {
	Homeserver string
}

// Validate checks that required fields are set.
func (c Config) Validate() error {
	if c.Homeserver == "" {
		return errors.New("matrix: homeserver is required")
	}
	return nil
}
