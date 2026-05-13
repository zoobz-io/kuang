//go:build testing

package github

import "testing"

func TestScopesReturnsAll(t *testing.T) {
	scopes := Scopes()

	expected := []string{
		ScopeReposRead, ScopeIssuesRead, ScopeIssuesWrite,
		ScopePullsRead, ScopePullsWrite, ScopeContentRead,
		ScopeContentWrite, ScopeSearchRead,
	}

	if len(scopes) != len(expected) {
		t.Fatalf("expected %d scopes, got %d", len(expected), len(scopes))
	}

	for i, s := range scopes {
		if s != expected[i] {
			t.Errorf("scope[%d] = %q, want %q", i, s, expected[i])
		}
	}
}

func TestScopesAreUnique(t *testing.T) {
	seen := make(map[string]bool)
	for _, s := range Scopes() {
		if seen[s] {
			t.Errorf("duplicate scope: %q", s)
		}
		seen[s] = true
	}
}
