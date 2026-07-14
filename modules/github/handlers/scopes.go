package handlers

// Scope constants for the GitHub module. These are used in kuang.yaml to
// grant agents access to specific operations. Patterns like "github-*-read"
// match all read scopes via the CLI's glob matching.
const (
	ScopeReposRead    = "github-repos-read"
	ScopeIssuesRead   = "github-issues-read"
	ScopeIssuesWrite  = "github-issues-write"
	ScopePullsRead    = "github-pulls-read"
	ScopePullsWrite   = "github-pulls-write"
	ScopeContentRead  = "github-content-read"
	ScopeContentWrite = "github-content-write"
	ScopeSearchRead   = "github-search-read"
)

// Scopes returns all scope strings defined by the GitHub module.
// Use this to set the server's permission ceiling in kuang.yaml or tests.
func Scopes() []string {
	return []string{
		ScopeReposRead, ScopeIssuesRead, ScopeIssuesWrite,
		ScopePullsRead, ScopePullsWrite, ScopeContentRead,
		ScopeContentWrite, ScopeSearchRead,
	}
}
