// Package services implements external service clients for kuang.
package services

import (
	"context"
	"fmt"

	"github.com/zoobz-io/kuang/types"
	"github.com/zoobz-io/kuang/config"
	"github.com/zoobz-io/kuang/internal/httpc"
)

// GitHub implements GitHub API operations via the instrumented HTTP client.
// All methods return api types ready for the handler to return directly.
type GitHub struct {
	client *httpc.Client
	owner  string
}

// NewGitHub creates a GitHub service from config.
func NewGitHub(cfg config.GitHub) *GitHub {
	client := httpc.New(
		httpc.WithBaseURL(cfg.APIURL),
		httpc.WithBearerToken(cfg.Token),
		httpc.WithHeader("Accept", "application/vnd.github+json"),
		httpc.WithHeader("X-GitHub-Api-Version", "2022-11-28"),
	)
	return &GitHub{client: client, owner: cfg.Owner}
}

// --- Repos ---

// ListRepos returns repositories for the configured owner.
func (g *GitHub) ListRepos(ctx context.Context) (types.RepoList, error) {
	resp, err := g.client.Get(ctx, fmt.Sprintf("/users/%s/repos", g.owner))
	if err != nil {
		return types.RepoList{}, err
	}
	var repos []types.Repo
	if err := resp.Decode(&repos); err != nil {
		return types.RepoList{}, err
	}
	return types.RepoList{Repos: repos}, nil
}

// GetRepo returns a single repository by name.
func (g *GitHub) GetRepo(ctx context.Context, name string) (types.Repo, error) {
	resp, err := g.client.Get(ctx, fmt.Sprintf("/repos/%s/%s", g.owner, name))
	if err != nil {
		return types.Repo{}, err
	}
	var repo types.Repo
	if err := resp.Decode(&repo); err != nil {
		return types.Repo{}, err
	}
	return repo, nil
}

// --- Issues ---

// ListIssues returns issues for a repository.
func (g *GitHub) ListIssues(ctx context.Context, repo string) (types.IssueList, error) {
	resp, err := g.client.Get(ctx, fmt.Sprintf("/repos/%s/%s/issues", g.owner, repo))
	if err != nil {
		return types.IssueList{}, err
	}
	var issues []types.Issue
	if err := resp.Decode(&issues); err != nil {
		return types.IssueList{}, err
	}
	return types.IssueList{Issues: issues}, nil
}

// GetIssue returns a single issue by number.
func (g *GitHub) GetIssue(ctx context.Context, repo string, number int) (types.Issue, error) {
	resp, err := g.client.Get(ctx, fmt.Sprintf("/repos/%s/%s/issues/%d", g.owner, repo, number))
	if err != nil {
		return types.Issue{}, err
	}
	var issue types.Issue
	if err := resp.Decode(&issue); err != nil {
		return types.Issue{}, err
	}
	return issue, nil
}

// CreateIssue creates a new issue in a repository.
func (g *GitHub) CreateIssue(ctx context.Context, repo, title, body string) (types.Issue, error) {
	resp, err := g.client.Post(ctx, fmt.Sprintf("/repos/%s/%s/issues", g.owner, repo), map[string]string{
		"title": title,
		"body":  body,
	})
	if err != nil {
		return types.Issue{}, err
	}
	var issue types.Issue
	if err := resp.Decode(&issue); err != nil {
		return types.Issue{}, err
	}
	return issue, nil
}

// --- Pull Requests ---

// ListPRs returns pull requests for a repository.
func (g *GitHub) ListPRs(ctx context.Context, repo string) (types.PRList, error) {
	resp, err := g.client.Get(ctx, fmt.Sprintf("/repos/%s/%s/pulls", g.owner, repo))
	if err != nil {
		return types.PRList{}, err
	}
	var prs []types.PullRequest
	if err := resp.Decode(&prs); err != nil {
		return types.PRList{}, err
	}
	return types.PRList{PullRequests: prs}, nil
}

// GetPR returns a single pull request by number.
func (g *GitHub) GetPR(ctx context.Context, repo string, number int) (types.PullRequest, error) {
	resp, err := g.client.Get(ctx, fmt.Sprintf("/repos/%s/%s/pulls/%d", g.owner, repo, number))
	if err != nil {
		return types.PullRequest{}, err
	}
	var pr types.PullRequest
	if err := resp.Decode(&pr); err != nil {
		return types.PullRequest{}, err
	}
	return pr, nil
}

// CreatePR creates a new pull request.
func (g *GitHub) CreatePR(ctx context.Context, repo, title, body, head, base string) (types.PullRequest, error) {
	resp, err := g.client.Post(ctx, fmt.Sprintf("/repos/%s/%s/pulls", g.owner, repo), map[string]string{
		"title": title,
		"body":  body,
		"head":  head,
		"base":  base,
	})
	if err != nil {
		return types.PullRequest{}, err
	}
	var pr types.PullRequest
	if err := resp.Decode(&pr); err != nil {
		return types.PullRequest{}, err
	}
	return pr, nil
}

// --- Content ---

// GetFile retrieves a file's content from a repository.
func (g *GitHub) GetFile(ctx context.Context, repo, path, ref string) (types.FileContent, error) {
	url := fmt.Sprintf("/repos/%s/%s/contents/%s", g.owner, repo, path)
	if ref != "" {
		url += "?ref=" + ref
	}
	resp, err := g.client.Get(ctx, url)
	if err != nil {
		return types.FileContent{}, err
	}
	var fc types.FileContent
	if err := resp.Decode(&fc); err != nil {
		return types.FileContent{}, err
	}
	return fc, nil
}

// CreateOrUpdateFile creates or updates a file in a repository.
func (g *GitHub) CreateOrUpdateFile(ctx context.Context, repo, path, message, content, sha string) (types.FileContent, error) {
	body := map[string]string{
		"message": message,
		"content": content,
	}
	if sha != "" {
		body["sha"] = sha
	}
	resp, err := g.client.Put(ctx, fmt.Sprintf("/repos/%s/%s/contents/%s", g.owner, repo, path), body)
	if err != nil {
		return types.FileContent{}, err
	}
	var fc types.FileContent
	if err := resp.Decode(&fc); err != nil {
		return types.FileContent{}, err
	}
	return fc, nil
}

// --- Search ---

// SearchCode searches for code across GitHub.
func (g *GitHub) SearchCode(ctx context.Context, query string) (types.CodeSearchResult, error) {
	resp, err := g.client.Get(ctx, "/search/code?q="+query)
	if err != nil {
		return types.CodeSearchResult{}, err
	}
	var result types.CodeSearchResult
	if err := resp.Decode(&result); err != nil {
		return types.CodeSearchResult{}, err
	}
	return result, nil
}
