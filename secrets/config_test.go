package secrets

import (
	"context"
	"testing"
)

type stubProvider struct {
	secrets map[string]string
}

func (s *stubProvider) Get(_ context.Context, key string) (string, error) {
	if v, ok := s.secrets[key]; ok {
		return v, nil
	}
	return "", context.Canceled
}

func TestConfigValidate(t *testing.T) {
	cfg := Config{
		Provider: &stubProvider{},
		Scopes:   map[string][]string{"secrets-db": {"db/password"}},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
}

func TestConfigValidateNilProvider(t *testing.T) {
	cfg := Config{Scopes: map[string][]string{"secrets-db": {"db/password"}}}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for nil provider")
	}
}

func TestConfigValidateNoScopes(t *testing.T) {
	cfg := Config{Provider: &stubProvider{}}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for empty scopes")
	}
}

func TestConfigValidateEmptyKeyList(t *testing.T) {
	cfg := Config{
		Provider: &stubProvider{},
		Scopes:   map[string][]string{"secrets-db": {}},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for scope with no keys")
	}
}
