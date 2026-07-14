package models

import "github.com/zoobz-io/check"

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
	Title string `json:"title"`
	Body  string `json:"body"`
	Head  string `json:"head"`
	Base  string `json:"base"`
}

// Validate validates a CreatePRRequest.
func (r CreatePRRequest) Validate() error {
	return check.All(
		check.Str(r.Title, "title").Required().V(),
		check.Str(r.Head, "head").Required().V(),
		check.Str(r.Base, "base").Required().V(),
	).Err()
}
