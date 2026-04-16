// Package config defines typed configuration structs loaded via fig.
package config

import "errors"

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
	if c.CACertPath == "" {
		return errors.New("CA cert path is required")
	}
	if c.CertPath == "" {
		return errors.New("server cert path is required")
	}
	if c.KeyPath == "" {
		return errors.New("server key path is required")
	}
	switch c.CryptoAlgo {
	case "ed25519", "ecdsa-p256":
	default:
		return errors.New("crypto algo must be ed25519 or ecdsa-p256")
	}
	return nil
}
