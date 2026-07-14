package mcp

import "testing"

func TestConfigValidate(t *testing.T) {
	valid := Config{
		Host: "https://localhost:8080",
		CA:   "certs/ca.pem",
		Cert: "certs/client.pem",
		Key:  "certs/client-key.pem",
	}

	tests := []struct {
		name    string
		mutate  func(c *Config)
		wantErr bool
	}{
		{"valid", func(*Config) {}, false},
		{"empty host", func(c *Config) { c.Host = "" }, true},
		{"empty ca", func(c *Config) { c.CA = "" }, true},
		{"empty cert", func(c *Config) { c.Cert = "" }, true},
		{"empty key", func(c *Config) { c.Key = "" }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := valid
			tt.mutate(&cfg)
			err := cfg.Validate()
			if tt.wantErr && err == nil {
				t.Errorf("Validate() = nil, want error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Validate() = %v, want nil", err)
			}
		})
	}
}
