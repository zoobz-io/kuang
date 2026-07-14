package handlers

import "testing"

func TestEndpointsCount(t *testing.T) {
	eps := Endpoints()
	if len(eps) != 13 {
		t.Errorf("endpoints = %d, want 13", len(eps))
	}
}

func TestScopeConstants(t *testing.T) {
	scopes := []string{
		ScopeRoomsRead,
		ScopeRoomsWrite,
		ScopeMessagesRead,
		ScopeMessagesWrite,
		ScopeMembersRead,
		ScopeInvitesRead,
		ScopeInvitesWrite,
		ScopeWatchRead,
		ScopeDMRead,
		ScopeDMWrite,
		ScopeIdentityRead,
	}
	seen := map[string]bool{}
	for _, s := range scopes {
		if s == "" {
			t.Error("empty scope constant")
		}
		if seen[s] {
			t.Errorf("duplicate scope: %s", s)
		}
		seen[s] = true
	}
}

// TestScopesCompleteness verifies Scopes() returns every scope constant
// exactly once, with no empties or duplicates.
func TestScopesCompleteness(t *testing.T) {
	want := []string{
		ScopeRoomsRead,
		ScopeRoomsWrite,
		ScopeMessagesRead,
		ScopeMessagesWrite,
		ScopeMembersRead,
		ScopeInvitesRead,
		ScopeInvitesWrite,
		ScopeWatchRead,
		ScopeDMRead,
		ScopeDMWrite,
		ScopeIdentityRead,
	}

	got := Scopes()
	if len(got) != len(want) {
		t.Fatalf("Scopes() len = %d, want %d", len(got), len(want))
	}

	seen := map[string]bool{}
	for _, s := range got {
		if s == "" {
			t.Error("Scopes() contains empty scope")
		}
		if seen[s] {
			t.Errorf("Scopes() contains duplicate scope: %s", s)
		}
		seen[s] = true
	}

	for _, w := range want {
		if !seen[w] {
			t.Errorf("Scopes() missing scope: %s", w)
		}
	}
}
