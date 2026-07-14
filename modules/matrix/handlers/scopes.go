package handlers

// Scope constants for the Matrix module. These are used in kuang.yaml to
// grant agents access to specific operations. Patterns like "matrix-*-read"
// match all read scopes via the CLI's glob matching.
const (
	ScopeRoomsRead     = "matrix-rooms-read"
	ScopeRoomsWrite    = "matrix-rooms-write"
	ScopeMessagesRead  = "matrix-messages-read"
	ScopeMessagesWrite = "matrix-messages-write"
	ScopeMembersRead   = "matrix-members-read"
	ScopeInvitesRead   = "matrix-invites-read"
	ScopeInvitesWrite  = "matrix-invites-write"
	ScopeWatchRead     = "matrix-watch-read"
	ScopeDMRead        = "matrix-dm-read"
	ScopeDMWrite       = "matrix-dm-write"
	ScopeIdentityRead  = "matrix-identity-read"
)

// Scopes returns all scope strings defined by the Matrix module.
// Use this to set the server's permission ceiling in kuang.yaml or tests.
func Scopes() []string {
	return []string{
		ScopeRoomsRead, ScopeRoomsWrite,
		ScopeMessagesRead, ScopeMessagesWrite,
		ScopeMembersRead,
		ScopeInvitesRead, ScopeInvitesWrite,
		ScopeWatchRead,
		ScopeDMRead, ScopeDMWrite,
		ScopeIdentityRead,
	}
}
