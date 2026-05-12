package github

import (
	"context"
	"fmt"

	"github.com/zoobz-io/kuang/httpc"
)

const defaultAPIURL = "https://api.github.com"

// service implements GitHub API operations via the instrumented HTTP client.
type service struct {
	client *httpc.Client
	owner  string
}

func newService(cfg Config) *service {
	apiURL := cfg.APIURL
	if apiURL == "" {
		apiURL = defaultAPIURL
	}
	client := httpc.New(
		httpc.WithBaseURL(apiURL),
		httpc.WithBearerToken(cfg.Token),
		httpc.WithHeader("Accept", "application/vnd.github+json"),
		httpc.WithHeader("X-GitHub-Api-Version", "2022-11-28"),
	)
	return &service{client: client, owner: cfg.Owner}
}

// --- Repos ---

func (s *service) ListRepos(ctx context.Context) (RepoList, error) {
	resp, err := s.client.Get(ctx, fmt.Sprintf("/users/%s/repos", s.owner))
	if err != nil {
		return RepoList{}, err
	}
	var repos []Repo
	if err := resp.Decode(&repos); err != nil {
		return RepoList{}, err
	}
	return RepoList{Repos: repos}, nil
}

func (s *service) GetRepo(ctx context.Context, name string) (Repo, error) {
	resp, err := s.client.Get(ctx, fmt.Sprintf("/repos/%s/%s", s.owner, name))
	if err != nil {
		return Repo{}, err
	}
	var repo Repo
	if err := resp.Decode(&repo); err != nil {
		return Repo{}, err
	}
	return repo, nil
}

// --- Issues ---

func (s *service) ListIssues(ctx context.Context, repo string) (IssueList, error) {
	resp, err := s.client.Get(ctx, fmt.Sprintf("/repos/%s/%s/issues", s.owner, repo))
	if err != nil {
		return IssueList{}, err
	}
	var issues []Issue
	if err := resp.Decode(&issues); err != nil {
		return IssueList{}, err
	}
	return IssueList{Issues: issues}, nil
}

func (s *service) GetIssue(ctx context.Context, repo string, number int) (Issue, error) {
	resp, err := s.client.Get(ctx, fmt.Sprintf("/repos/%s/%s/issues/%d", s.owner, repo, number))
	if err != nil {
		return Issue{}, err
	}
	var issue Issue
	if err := resp.Decode(&issue); err != nil {
		return Issue{}, err
	}
	return issue, nil
}

func (s *service) CreateIssue(ctx context.Context, repo, title, body string) (Issue, error) {
	resp, err := s.client.Post(ctx, fmt.Sprintf("/repos/%s/%s/issues", s.owner, repo), map[string]string{
		"title": title,
		"body":  body,
	})
	if err != nil {
		return Issue{}, err
	}
	var issue Issue
	if err := resp.Decode(&issue); err != nil {
		return Issue{}, err
	}
	return issue, nil
}

// --- Pull Requests ---

func (s *service) ListPRs(ctx context.Context, repo string) (PRList, error) {
	resp, err := s.client.Get(ctx, fmt.Sprintf("/repos/%s/%s/pulls", s.owner, repo))
	if err != nil {
		return PRList{}, err
	}
	var prs []PullRequest
	if err := resp.Decode(&prs); err != nil {
		return PRList{}, err
	}
	return PRList{PullRequests: prs}, nil
}

func (s *service) GetPR(ctx context.Context, repo string, number int) (PullRequest, error) {
	resp, err := s.client.Get(ctx, fmt.Sprintf("/repos/%s/%s/pulls/%d", s.owner, repo, number))
	if err != nil {
		return PullRequest{}, err
	}
	var pr PullRequest
	if err := resp.Decode(&pr); err != nil {
		return PullRequest{}, err
	}
	return pr, nil
}

func (s *service) CreatePR(ctx context.Context, repo, title, body, head, base string) (PullRequest, error) {
	resp, err := s.client.Post(ctx, fmt.Sprintf("/repos/%s/%s/pulls", s.owner, repo), map[string]string{
		"title": title,
		"body":  body,
		"head":  head,
		"base":  base,
	})
	if err != nil {
		return PullRequest{}, err
	}
	var pr PullRequest
	if err := resp.Decode(&pr); err != nil {
		return PullRequest{}, err
	}
	return pr, nil
}

// --- Content ---

func (s *service) GetFile(ctx context.Context, repo, path, ref string) (FileContent, error) {
	url := fmt.Sprintf("/repos/%s/%s/contents/%s", s.owner, repo, path)
	if ref != "" {
		url += "?ref=" + ref
	}
	resp, err := s.client.Get(ctx, url)
	if err != nil {
		return FileContent{}, err
	}
	var fc FileContent
	if err := resp.Decode(&fc); err != nil {
		return FileContent{}, err
	}
	return fc, nil
}

func (s *service) CreateOrUpdateFile(ctx context.Context, repo, path, message, content, sha string) (FileContent, error) {
	body := map[string]string{
		"message": message,
		"content": content,
	}
	if sha != "" {
		body["sha"] = sha
	}
	resp, err := s.client.Put(ctx, fmt.Sprintf("/repos/%s/%s/contents/%s", s.owner, repo, path), body)
	if err != nil {
		return FileContent{}, err
	}
	var fc FileContent
	if err := resp.Decode(&fc); err != nil {
		return FileContent{}, err
	}
	return fc, nil
}

// --- Search ---

func (s *service) SearchCode(ctx context.Context, query string) (CodeSearchResult, error) {
	resp, err := s.client.Get(ctx, "/search/code?q="+query)
	if err != nil {
		return CodeSearchResult{}, err
	}
	var result CodeSearchResult
	if err := resp.Decode(&result); err != nil {
		return CodeSearchResult{}, err
	}
	return result, nil
}
