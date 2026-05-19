package secrets

import "testing"

func TestEndpointsCount(t *testing.T) {
	svc := &service{provider: &stubProvider{}}
	eps := endpoints(svc, []string{"secrets-db"})
	if len(eps) != 1 {
		t.Errorf("endpoints = %d, want 1", len(eps))
	}
}

func TestScopesFromConfig(t *testing.T) {
	cfg := Config{
		Provider: &stubProvider{},
		Scopes: map[string][]string{
			"secrets-db":  {"db/password"},
			"secrets-api": {"api/token"},
		},
	}
	scopes := Scopes(cfg)
	if len(scopes) != 2 {
		t.Fatalf("scopes = %d, want 2", len(scopes))
	}
	// Sorted.
	if scopes[0] != "secrets-api" || scopes[1] != "secrets-db" {
		t.Errorf("scopes = %v, want [secrets-api secrets-db]", scopes)
	}
}

func TestScopesAreUnique(t *testing.T) {
	cfg := Config{
		Provider: &stubProvider{},
		Scopes: map[string][]string{
			"secrets-db":  {"db/password", "db/host"},
			"secrets-api": {"api/token"},
		},
	}
	seen := make(map[string]bool)
	for _, s := range Scopes(cfg) {
		if seen[s] {
			t.Errorf("duplicate scope: %q", s)
		}
		seen[s] = true
	}
}
