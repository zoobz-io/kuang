package secrets

import (
	"context"
	"strings"
	"testing"
)

// stubIdentity implements rocco.Identity for testing.
type stubIdentity struct {
	scopes []string
}

func (s *stubIdentity) ID() string                { return "test-agent" }
func (s *stubIdentity) TenantID() string           { return "test" }
func (s *stubIdentity) Email() string              { return "" }
func (s *stubIdentity) Scopes() []string           { return s.scopes }
func (s *stubIdentity) Roles() []string            { return nil }
func (s *stubIdentity) HasScope(scope string) bool {
	for _, s2 := range s.scopes {
		if s2 == scope {
			return true
		}
	}
	return false
}
func (s *stubIdentity) HasRole(_ string) bool  { return false }
func (s *stubIdentity) Stats() map[string]int  { return nil }

func TestServiceGetAuthorized(t *testing.T) {
	svc := newService(Config{
		Provider: &stubProvider{secrets: map[string]string{
			"db/password": "hunter2",
		}},
		Scopes: map[string][]string{
			"secrets-db": {"db/password"},
		},
	})

	identity := &stubIdentity{scopes: []string{"secrets-db"}}
	secret, err := svc.Get(context.Background(), identity, "db/password")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if secret.Key != "db/password" {
		t.Errorf("key = %q, want db/password", secret.Key)
	}
	if secret.Value != "hunter2" {
		t.Errorf("value = %q, want hunter2", secret.Value)
	}
}

func TestServiceGetDeniedWrongScope(t *testing.T) {
	svc := newService(Config{
		Provider: &stubProvider{secrets: map[string]string{
			"db/password": "hunter2",
		}},
		Scopes: map[string][]string{
			"secrets-db": {"db/password"},
		},
	})

	identity := &stubIdentity{scopes: []string{"secrets-api"}}
	_, err := svc.Get(context.Background(), identity, "db/password")
	if err == nil {
		t.Fatal("expected access denied")
	}
	if !strings.Contains(err.Error(), "access denied") {
		t.Errorf("error = %q, want access denied", err.Error())
	}
}

func TestServiceGetDeniedUnmappedKey(t *testing.T) {
	svc := newService(Config{
		Provider: &stubProvider{secrets: map[string]string{
			"db/password": "hunter2",
		}},
		Scopes: map[string][]string{
			"secrets-db": {"db/password"},
		},
	})

	identity := &stubIdentity{scopes: []string{"secrets-db"}}
	_, err := svc.Get(context.Background(), identity, "api/token")
	if err == nil {
		t.Fatal("expected access denied for unmapped key")
	}
}

func TestServiceGetMultipleScopes(t *testing.T) {
	svc := newService(Config{
		Provider: &stubProvider{secrets: map[string]string{
			"shared/key": "value",
		}},
		Scopes: map[string][]string{
			"secrets-db":  {"shared/key"},
			"secrets-api": {"shared/key"},
		},
	})

	// Either scope should work.
	for _, scope := range []string{"secrets-db", "secrets-api"} {
		identity := &stubIdentity{scopes: []string{scope}}
		secret, err := svc.Get(context.Background(), identity, "shared/key")
		if err != nil {
			t.Errorf("scope %q: unexpected error: %v", scope, err)
		}
		if secret.Value != "value" {
			t.Errorf("scope %q: value = %q", scope, secret.Value)
		}
	}
}

func TestServiceGetProviderError(t *testing.T) {
	svc := newService(Config{
		Provider: &stubProvider{secrets: map[string]string{}},
		Scopes: map[string][]string{
			"secrets-db": {"missing/key"},
		},
	})

	identity := &stubIdentity{scopes: []string{"secrets-db"}}
	_, err := svc.Get(context.Background(), identity, "missing/key")
	if err == nil {
		t.Fatal("expected error from provider")
	}
}
