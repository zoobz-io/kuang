// Package contracts defines the interfaces for the GitHub module's services.
package contracts

import (
	"context"

	"github.com/zoobzio/kuang/internal/httpc"
	"github.com/zoobzio/kuang/modules/github/models"
)

// API describes the GitHub REST API operations backing the module's endpoints.
// Handlers depend on this interface so the concrete service can be swapped or
// mocked in tests.
type API interface {
	// Repos.
	ListRepos(ctx context.Context, opts ...httpc.RequestOption) (models.RepoList, error)
	GetRepo(ctx context.Context, name string, opts ...httpc.RequestOption) (models.Repo, error)

	// Issues.
	ListIssues(ctx context.Context, repo string, opts ...httpc.RequestOption) (models.IssueList, error)
	GetIssue(ctx context.Context, repo string, number int, opts ...httpc.RequestOption) (models.Issue, error)
	CreateIssue(ctx context.Context, repo, title, body string, opts ...httpc.RequestOption) (models.Issue, error)

	// Pull requests.
	ListPRs(ctx context.Context, repo string, opts ...httpc.RequestOption) (models.PRList, error)
	GetPR(ctx context.Context, repo string, number int, opts ...httpc.RequestOption) (models.PullRequest, error)
	CreatePR(ctx context.Context, repo, title, body, head, base string, opts ...httpc.RequestOption) (models.PullRequest, error)

	// Content.
	GetFile(ctx context.Context, repo, path, ref string, opts ...httpc.RequestOption) (models.FileContent, error)
	CreateOrUpdateFile(ctx context.Context, repo, path, message, content, sha string, opts ...httpc.RequestOption) (models.FileContent, error)

	// Search.
	SearchCode(ctx context.Context, query string, opts ...httpc.RequestOption) (models.CodeSearchResult, error)
}
