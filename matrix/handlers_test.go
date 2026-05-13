package matrix

import "testing"

func TestEndpointsCount(t *testing.T) {
	svc := &service{homeserver: "https://matrix.localhost"}
	eps := endpoints(svc)
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
