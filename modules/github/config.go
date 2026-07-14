package github

import "github.com/zoobz-io/check"

// Config holds settings for connecting to the GitHub REST API.
type Config struct {
	APIURL string `env:"KUANG_GITHUB_API_URL" default:"https://api.github.com"`
	Owner  string `env:"KUANG_GITHUB_OWNER"`
}

// Validate checks that required fields are set and the API URL is well-formed.
func (a Config) Validate() error {
	return check.All(
		check.Str(a.APIURL, "api_url").Required().URL().V(),
		check.Str(a.Owner, "owner").Required().V(),
	).Err()
}
