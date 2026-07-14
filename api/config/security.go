// Package config defines typed configuration structs loaded via fig.
package config

import "github.com/zoobz-io/check"

// Security holds TLS certificate paths and cryptography settings.
type Security struct {
	CACertPath  string `env:"APP_CA_CERT_PATH" default:"certs/ca.pem"`
	CertPath    string `env:"APP_CERT_PATH" default:"certs/server.pem"`
	KeyPath     string `env:"APP_KEY_PATH" default:"certs/server-key.pem"`
	CryptoAlgo  string `env:"APP_CRYPTO_ALGO" default:"ed25519"`
	RequireMTLS bool   `env:"APP_REQUIRE_MTLS" default:"true"`
}

// Validate checks that all required fields are set and the crypto algo is supported.
func (c Security) Validate() error {
	return check.All(
		check.Str(c.CACertPath, "ca_cert_path").Required().V(),
		check.Str(c.CertPath, "cert_path").Required().V(),
		check.Str(c.KeyPath, "key_path").Required().V(),
		check.OneOfValues(c.CryptoAlgo, []string{"ed25519", "ecdsa-p256"}, "crypto_algo"),
	).Err()
}
