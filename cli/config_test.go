package cli

import (
	"os"
	"path/filepath"
	"testing"
)

const testConfig = `
server:
  port: 9090
  host: 0.0.0.0

modules:
  - weather
  - inventory

scopes:
  - weather-forecast-read
  - weather-forecast-write
  - weather-alerts-read
  - weather-alerts-admin
  - inventory-items-read
  - inventory-items-write
  - inventory-items-delete

agents:
  claude:
    scopes:
      - "weather-*-read"
      - "inventory-items-read"
      - "inventory-items-write"
  reader:
    scopes:
      - "*-*-read"
  admin:
    scopes:
      - "*"
`

func writeConfig(t *testing.T, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, "kuang.yaml")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	path := writeConfig(t, dir, testConfig)

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	if cfg.Server.Port != 9090 {
		t.Errorf("port = %d, want 9090", cfg.Server.Port)
	}
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("host = %q, want 0.0.0.0", cfg.Server.Host)
	}
	if len(cfg.Modules) != 2 {
		t.Errorf("modules = %d, want 2", len(cfg.Modules))
	}
	if len(cfg.Scopes) != 7 {
		t.Errorf("scopes = %d, want 7", len(cfg.Scopes))
	}
	if len(cfg.Agents) != 3 {
		t.Errorf("agents = %d, want 3", len(cfg.Agents))
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	dir := t.TempDir()
	path := writeConfig(t, dir, "scopes: []")

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("default port = %d, want 8080", cfg.Server.Port)
	}
	if cfg.Server.Host != "localhost" {
		t.Errorf("default host = %q, want localhost", cfg.Server.Host)
	}
}

func TestLoadConfigMissing(t *testing.T) {
	_, err := LoadConfig("/nonexistent/kuang.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestFindConfig(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "scopes: []")

	original, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(original) })

	path, err := FindConfig()
	if err != nil {
		t.Fatalf("FindConfig: %v", err)
	}
	if filepath.Base(path) != "kuang.yaml" {
		t.Errorf("found %q", path)
	}
}

func TestFindConfigWalksUp(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "scopes: []")

	sub := filepath.Join(dir, "a", "b", "c")
	if err := os.MkdirAll(sub, 0750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	original, _ := os.Getwd()
	if err := os.Chdir(sub); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(original) })

	path, err := FindConfig()
	if err != nil {
		t.Fatalf("FindConfig: %v", err)
	}
	if filepath.Dir(path) != dir {
		t.Errorf("expected config in %s, got %s", dir, filepath.Dir(path))
	}
}

func TestResolveAgentScopes(t *testing.T) {
	dir := t.TempDir()
	path := writeConfig(t, dir, testConfig)
	cfg, _ := LoadConfig(path)

	tests := []struct {
		agent string
		want  int
	}{
		{"claude", 4},  // weather-forecast-read, weather-alerts-read, inventory-items-read, inventory-items-write
		{"reader", 3},  // weather-forecast-read, weather-alerts-read, inventory-items-read
		{"admin", 7},   // all scopes
	}

	for _, tt := range tests {
		scopes, err := cfg.ResolveAgentScopes(tt.agent)
		if err != nil {
			t.Errorf("resolve %s: %v", tt.agent, err)
			continue
		}
		if len(scopes) != tt.want {
			t.Errorf("%s: got %d scopes %v, want %d", tt.agent, len(scopes), scopes, tt.want)
		}
	}
}

func TestResolveAgentScopesUnknownAgent(t *testing.T) {
	dir := t.TempDir()
	path := writeConfig(t, dir, testConfig)
	cfg, _ := LoadConfig(path)

	_, err := cfg.ResolveAgentScopes("ghost")
	if err == nil {
		t.Fatal("expected error for unknown agent")
	}
}

func TestResolveAgentScopesNoMatch(t *testing.T) {
	dir := t.TempDir()
	path := writeConfig(t, dir, `
scopes:
  - weather-forecast-read
agents:
  bad:
    scopes:
      - "nonexistent-*-read"
`)
	cfg, _ := LoadConfig(path)

	_, err := cfg.ResolveAgentScopes("bad")
	if err == nil {
		t.Fatal("expected error when pattern matches nothing")
	}
}

func TestGlobMatch(t *testing.T) {
	tests := []struct {
		pattern, value string
		want           bool
	}{
		{"weather-*-read", "weather-forecast-read", true},
		{"weather-*-read", "weather-alerts-read", true},
		{"weather-*-read", "weather-forecast-write", false},
		{"*-*-read", "weather-forecast-read", true},
		{"*-*-read", "inventory-items-read", true},
		{"*-*-read", "inventory-items-write", false},
		{"weather-forecast-read", "weather-forecast-read", true},
		{"weather-forecast-read", "weather-forecast-write", false},
		// Different segment counts don't match.
		{"*-read", "weather-forecast-read", false},
		{"*-*-*-read", "weather-forecast-read", false},
	}

	for _, tt := range tests {
		got := globMatch(tt.pattern, tt.value)
		if got != tt.want {
			t.Errorf("globMatch(%q, %q) = %v, want %v", tt.pattern, tt.value, got, tt.want)
		}
	}
}

func TestDedupeScopes(t *testing.T) {
	dir := t.TempDir()
	path := writeConfig(t, dir, `
scopes:
  - weather-forecast-read
  - inventory-items-read
agents:
  overlap:
    scopes:
      - "*-*-read"
      - "weather-forecast-read"
`)
	cfg, _ := LoadConfig(path)

	scopes, err := cfg.ResolveAgentScopes("overlap")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if len(scopes) != 2 {
		t.Errorf("expected 2 deduped scopes, got %d: %v", len(scopes), scopes)
	}
}
