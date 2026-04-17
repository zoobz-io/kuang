// Package contracts defines service interfaces for kuang.
package contracts

import (
	"context"

	"github.com/zoobz-io/kuang/types"
)

// GitHub defines all GitHub API operations.
type GitHub interface {
	ListRepos(ctx context.Context) (types.RepoList, error)
	GetRepo(ctx context.Context, name string) (types.Repo, error)
	ListIssues(ctx context.Context, repo string) (types.IssueList, error)
	GetIssue(ctx context.Context, repo string, number int) (types.Issue, error)
	CreateIssue(ctx context.Context, repo, title, body string) (types.Issue, error)
	ListPRs(ctx context.Context, repo string) (types.PRList, error)
	GetPR(ctx context.Context, repo string, number int) (types.PullRequest, error)
	CreatePR(ctx context.Context, repo, title, body, head, base string) (types.PullRequest, error)
	GetFile(ctx context.Context, repo, path, ref string) (types.FileContent, error)
	CreateOrUpdateFile(ctx context.Context, repo, path, message, content, sha string) (types.FileContent, error)
	SearchCode(ctx context.Context, query string) (types.CodeSearchResult, error)
}
