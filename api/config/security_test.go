package config

import "testing"

func TestSecurityValidate(t *testing.T) {
	valid := Security{
		CACertPath: "ca.pem",
		CertPath:   "server.pem",
		KeyPath:    "server-key.pem",
		CryptoAlgo: "ed25519",
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
}

func TestSecurityValidateECDSA(t *testing.T) {
	cfg := Security{
		CACertPath: "ca.pem",
		CertPath:   "server.pem",
		KeyPath:    "server-key.pem",
		CryptoAlgo: "ecdsa-p256",
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid for ecdsa-p256, got %v", err)
	}
}

func TestSecurityValidateEmptyCACertPath(t *testing.T) {
	cfg := Security{CertPath: "server.pem", KeyPath: "key.pem", CryptoAlgo: "ed25519"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for empty CA cert path")
	}
}

func TestSecurityValidateEmptyCertPath(t *testing.T) {
	cfg := Security{CACertPath: "ca.pem", KeyPath: "key.pem", CryptoAlgo: "ed25519"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for empty cert path")
	}
}

func TestSecurityValidateEmptyKeyPath(t *testing.T) {
	cfg := Security{CACertPath: "ca.pem", CertPath: "server.pem", CryptoAlgo: "ed25519"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for empty key path")
	}
}

func TestSecurityValidateInvalidCryptoAlgo(t *testing.T) {
	cfg := Security{
		CACertPath: "ca.pem",
		CertPath:   "server.pem",
		KeyPath:    "key.pem",
		CryptoAlgo: "rsa",
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for invalid crypto algo")
	}
}
