package github

import (
	"context"
	"fmt"

	"github.com/zoobzio/kuang/internal/httpc"
	"github.com/zoobzio/kuang/modules/github/models"
)

// API implements GitHub API operations via the instrumented HTTP client.
type API struct {
	client *httpc.Client
	owner  string
}

// NewAPI constructs an API service from the given config.
func NewAPI(cfg Config) *API {
	client := httpc.New(
		httpc.WithBaseURL(cfg.APIURL),
		httpc.WithHeader("Accept", "application/vnd.github+json"),
		httpc.WithHeader("X-GitHub-Api-Version", "2022-11-28"),
	)
	return &API{client: client, owner: cfg.Owner}
}

// --- Repos ---

// ListRepos returns the configured owner's repositories.
func (s *API) ListRepos(ctx context.Context, opts ...httpc.RequestOption) (models.RepoList, error) {
	resp, err := s.client.Get(ctx, fmt.Sprintf("/users/%s/repos", s.owner), opts...)
	if err != nil {
		return models.RepoList{}, err
	}
	var repos []models.Repo
	if err := resp.Decode(&repos); err != nil {
		return models.RepoList{}, err
	}
	return models.RepoList{Repos: repos}, nil
}

// GetRepo returns a single repository owned by the configured owner.
func (s *API) GetRepo(ctx context.Context, name string, opts ...httpc.RequestOption) (models.Repo, error) {
	resp, err := s.client.Get(ctx, fmt.Sprintf("/repos/%s/%s", s.owner, name), opts...)
	if err != nil {
		return models.Repo{}, err
	}
	var repo models.Repo
	if err := resp.Decode(&repo); err != nil {
		return models.Repo{}, err
	}
	return repo, nil
}

// --- Issues ---

// ListIssues returns the issues for a repository.
func (s *API) ListIssues(ctx context.Context, repo string, opts ...httpc.RequestOption) (models.IssueList, error) {
	resp, err := s.client.Get(ctx, fmt.Sprintf("/repos/%s/%s/issues", s.owner, repo), opts...)
	if err != nil {
		return models.IssueList{}, err
	}
	var issues []models.Issue
	if err := resp.Decode(&issues); err != nil {
		return models.IssueList{}, err
	}
	return models.IssueList{Issues: issues}, nil
}

// GetIssue returns a single issue by number.
func (s *API) GetIssue(ctx context.Context, repo string, number int, opts ...httpc.RequestOption) (models.Issue, error) {
	resp, err := s.client.Get(ctx, fmt.Sprintf("/repos/%s/%s/issues/%d", s.owner, repo, number), opts...)
	if err != nil {
		return models.Issue{}, err
	}
	var issue models.Issue
	if err := resp.Decode(&issue); err != nil {
		return models.Issue{}, err
	}
	return issue, nil
}

// CreateIssue opens a new issue on a repository.
func (s *API) CreateIssue(ctx context.Context, repo, title, body string, opts ...httpc.RequestOption) (models.Issue, error) {
	resp, err := s.client.Post(ctx, fmt.Sprintf("/repos/%s/%s/issues", s.owner, repo), map[string]string{
		"title": title,
		"body":  body,
	}, opts...)
	if err != nil {
		return models.Issue{}, err
	}
	var issue models.Issue
	if err := resp.Decode(&issue); err != nil {
		return models.Issue{}, err
	}
	return issue, nil
}

// --- Pull Requests ---

// ListPRs returns the pull requests for a repository.
func (s *API) ListPRs(ctx context.Context, repo string, opts ...httpc.RequestOption) (models.PRList, error) {
	resp, err := s.client.Get(ctx, fmt.Sprintf("/repos/%s/%s/pulls", s.owner, repo), opts...)
	if err != nil {
		return models.PRList{}, err
	}
	var prs []models.PullRequest
	if err := resp.Decode(&prs); err != nil {
		return models.PRList{}, err
	}
	return models.PRList{PullRequests: prs}, nil
}

// GetPR returns a single pull request by number.
func (s *API) GetPR(ctx context.Context, repo string, number int, opts ...httpc.RequestOption) (models.PullRequest, error) {
	resp, err := s.client.Get(ctx, fmt.Sprintf("/repos/%s/%s/pulls/%d", s.owner, repo, number), opts...)
	if err != nil {
		return models.PullRequest{}, err
	}
	var pr models.PullRequest
	if err := resp.Decode(&pr); err != nil {
		return models.PullRequest{}, err
	}
	return pr, nil
}

// CreatePR opens a new pull request from head into base.
func (s *API) CreatePR(ctx context.Context, repo, title, body, head, base string, opts ...httpc.RequestOption) (models.PullRequest, error) {
	resp, err := s.client.Post(ctx, fmt.Sprintf("/repos/%s/%s/pulls", s.owner, repo), map[string]string{
		"title": title,
		"body":  body,
		"head":  head,
		"base":  base,
	}, opts...)
	if err != nil {
		return models.PullRequest{}, err
	}
	var pr models.PullRequest
	if err := resp.Decode(&pr); err != nil {
		return models.PullRequest{}, err
	}
	return pr, nil
}

// --- Content ---

// GetFile returns the contents of a file, optionally at a specific ref.
func (s *API) GetFile(ctx context.Context, repo, path, ref string, opts ...httpc.RequestOption) (models.FileContent, error) {
	url := fmt.Sprintf("/repos/%s/%s/contents/%s", s.owner, repo, path)
	if ref != "" {
		url += "?ref=" + ref
	}
	resp, err := s.client.Get(ctx, url, opts...)
	if err != nil {
		return models.FileContent{}, err
	}
	var fc models.FileContent
	if err := resp.Decode(&fc); err != nil {
		return models.FileContent{}, err
	}
	return fc, nil
}

// CreateOrUpdateFile writes a file. A non-empty sha updates an existing file;
// an empty sha creates a new one.
func (s *API) CreateOrUpdateFile(ctx context.Context, repo, path, message, content, sha string, opts ...httpc.RequestOption) (models.FileContent, error) {
	body := map[string]string{
		"message": message,
		"content": content,
	}
	if sha != "" {
		body["sha"] = sha
	}
	resp, err := s.client.Put(ctx, fmt.Sprintf("/repos/%s/%s/contents/%s", s.owner, repo, path), body, opts...)
	if err != nil {
		return models.FileContent{}, err
	}
	var fc models.FileContent
	if err := resp.Decode(&fc); err != nil {
		return models.FileContent{}, err
	}
	return fc, nil
}

// --- Search ---

// SearchCode runs a code search using GitHub's search query syntax.
func (s *API) SearchCode(ctx context.Context, query string, opts ...httpc.RequestOption) (models.CodeSearchResult, error) {
	resp, err := s.client.Get(ctx, "/search/code?q="+query, opts...)
	if err != nil {
		return models.CodeSearchResult{}, err
	}
	var result models.CodeSearchResult
	if err := resp.Decode(&result); err != nil {
		return models.CodeSearchResult{}, err
	}
	return result, nil
}
