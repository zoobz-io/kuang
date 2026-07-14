package config

import "testing"

func TestServerValidate(t *testing.T) {
	valid := Server{
		Host:   "localhost",
		DBPath: "data/kuang.db",
		Port:   8080,
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
}

func TestServerValidateEmptyHost(t *testing.T) {
	cfg := Server{DBPath: "data/kuang.db", Port: 8080}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for empty host")
	}
}

func TestServerValidateEmptyDBPath(t *testing.T) {
	cfg := Server{Host: "localhost", Port: 8080}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for empty db_path")
	}
}

func TestServerValidateBadPort(t *testing.T) {
	tests := []struct {
		name string
		port int
	}{
		{"zero", 0},
		{"negative", -1},
		{"too large", 70000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Server{Host: "localhost", DBPath: "data/kuang.db", Port: tt.port}
			if err := cfg.Validate(); err == nil {
				t.Fatalf("expected error for port %d", tt.port)
			}
		})
	}
}
