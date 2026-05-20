// Package github provides a kuang module for GitHub API operations.
package github

import "errors"

// Config holds GitHub API settings.
type Config struct {
	APIURL string
	Owner  string
}

// Validate checks that required fields are set.
func (c Config) Validate() error {
	if c.Owner == "" {
		return errors.New("github: owner is required")
	}
	return nil
}
