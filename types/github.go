// Package types defines request and response types for the kuang API boundary.
package types //nolint:revive // intentional package name

import "github.com/zoobz-io/check"

// --- Repos ---

// Repo is the API representation of a GitHub repository.
type Repo struct {
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	Description   string `json:"description"`
	DefaultBranch string `json:"default_branch"`
	HTMLURL       string `json:"html_url"`
	CloneURL      string `json:"clone_url"`
	Language      string `json:"language"`
	Stars         int    `json:"stargazers_count"`
	OpenIssues    int    `json:"open_issues_count"`
	Private       bool   `json:"private"`
	Fork          bool   `json:"fork"`
}

// Validate validates a Repo.
func (r Repo) Validate() error {
	return check.All(
		check.Str(r.Name, "name").Required().V(),
	).Err()
}

// RepoList is the response for listing repositories.
type RepoList struct {
	Repos []Repo `json:"repos"`
}

// Validate validates a RepoList.
func (r RepoList) Validate() error { return nil }

// --- Issues ---

// Issue is the API representation of a GitHub issue.
type Issue struct {
	Title   string  `json:"title"`
	Body    string  `json:"body"`
	State   string  `json:"state"`
	HTMLURL string  `json:"html_url"`
	Labels  []Label `json:"labels"`
	Number  int     `json:"number"`
}

// Label is a GitHub label.
type Label struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// Validate validates an Issue.
func (i Issue) Validate() error {
	return check.All(
		check.Str(i.Title, "title").Required().V(),
	).Err()
}

// IssueList is the response for listing issues.
type IssueList struct {
	Issues []Issue `json:"issues"`
}

// Validate validates an IssueList.
func (i IssueList) Validate() error { return nil }

// CreateIssueRequest is the request body for creating an issue.
type CreateIssueRequest struct {
	Title string `json:"title" validate:"required"`
	Body  string `json:"body"`
}

// Validate validates a CreateIssueRequest.
func (r CreateIssueRequest) Validate() error {
	return check.Check[CreateIssueRequest](
		check.Str(r.Title, "title").Required().V(),
	).Err()
}

// --- Pull Requests ---

// PullRequest is the API representation of a GitHub pull request.
type PullRequest struct {
	Title   string      `json:"title"`
	Body    string      `json:"body"`
	State   string      `json:"state"`
	HTMLURL string      `json:"html_url"`
	Head    PRBranchRef `json:"head"`
	Base    PRBranchRef `json:"base"`
	Number  int         `json:"number"`
	Draft   bool        `json:"draft"`
	Merged  bool        `json:"merged"`
}

// PRBranchRef is a branch reference on a pull request.
type PRBranchRef struct {
	Ref string `json:"ref"`
	SHA string `json:"sha"`
}

// Validate validates a PullRequest.
func (p PullRequest) Validate() error {
	return check.All(
		check.Str(p.Title, "title").Required().V(),
	).Err()
}

// PRList is the response for listing pull requests.
type PRList struct {
	PullRequests []PullRequest `json:"pull_requests"`
}

// Validate validates a PRList.
func (p PRList) Validate() error { return nil }

// CreatePRRequest is the request body for creating a pull request.
type CreatePRRequest struct {
	Title string `json:"title" validate:"required"`
	Body  string `json:"body"`
	Head  string `json:"head" validate:"required"`
	Base  string `json:"base" validate:"required"`
}

// Validate validates a CreatePRRequest.
func (r CreatePRRequest) Validate() error {
	return check.Check[CreatePRRequest](
		check.Str(r.Title, "title").Required().V(),
		check.Str(r.Head, "head").Required().V(),
		check.Str(r.Base, "base").Required().V(),
	).Err()
}

// --- Content ---

// FileContent is the API representation of a file in a repository.
type FileContent struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	SHA     string `json:"sha"`
}

// Validate validates a FileContent.
func (f FileContent) Validate() error {
	return check.All(
		check.Str(f.Path, "path").Required().V(),
	).Err()
}

// CreateOrUpdateFileRequest is the request body for creating/updating a file.
type CreateOrUpdateFileRequest struct {
	Message string `json:"message" validate:"required"`
	Content string `json:"content" validate:"required"`
	SHA     string `json:"sha"`
}

// Validate validates a CreateOrUpdateFileRequest.
func (r CreateOrUpdateFileRequest) Validate() error {
	return check.Check[CreateOrUpdateFileRequest](
		check.Str(r.Message, "message").Required().V(),
		check.Str(r.Content, "content").Required().V(),
	).Err()
}

// --- Search ---

// SearchResult is a single code search match.
type SearchResult struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	HTMLURL    string `json:"html_url"`
	Repository Repo   `json:"repository"`
}

// Validate validates a SearchResult.
func (s SearchResult) Validate() error { return nil }

// CodeSearchResult is the response for code search.
// GitHub returns items in the "items" field, but we expose as "results".
type CodeSearchResult struct {
	Results []SearchResult `json:"items"`
}

// Validate validates a CodeSearchResult.
func (c CodeSearchResult) Validate() error { return nil }
