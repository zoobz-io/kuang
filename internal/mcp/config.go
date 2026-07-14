package mcp

import "github.com/zoobz-io/check"

// Config holds the endpoint address and client certificate paths an agent uses to
// connect to kuang over mTLS.
type Config struct {
	Host string `env:"KUANG_API_HOST"`
	CA   string `env:"KUANG_API_CA_CERT_PATH" default:"certs/ca.pem"`
	Cert string `env:"KUANG_API_CERT_PATH" default:"certs/client.pem"`
	Key  string `env:"KUANG_API_KEY_PATH" default:"certs/client-key.pem"`
}

// Validate checks that all required fields are set.
func (c Config) Validate() error {
	return check.All(
		check.Str(c.Host, "host").Required().V(),
		check.Str(c.CA, "ca").Required().V(),
		check.Str(c.Cert, "cert").Required().V(),
		check.Str(c.Key, "key").Required().V(),
	).Err()
}
