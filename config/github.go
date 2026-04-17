package config

import "errors"

// GitHub holds GitHub API settings.
type GitHub struct {
	Token  string `env:"GITHUB_TOKEN" secret:"github/token"`
	APIURL string `env:"GITHUB_API_URL" default:"https://api.github.com"`
	Owner  string `env:"GITHUB_OWNER"`
}

// Validate checks that required fields are set.
func (c GitHub) Validate() error {
	if c.Token == "" {
		return errors.New("github token is required")
	}
	if c.Owner == "" {
		return errors.New("github owner is required")
	}
	return nil
}
