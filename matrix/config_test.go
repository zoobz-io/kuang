package matrix

import "testing"

func TestConfigValidate(t *testing.T) {
	cfg := Config{Homeserver: "https://matrix.example.com"}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
}

func TestConfigValidateEmptyHomeserver(t *testing.T) {
	cfg := Config{}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for empty homeserver")
	}
}
