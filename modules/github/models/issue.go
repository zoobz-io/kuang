package models

import "github.com/zoobz-io/check"

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
	Title string `json:"title"`
	Body  string `json:"body"`
}

// Validate validates a CreateIssueRequest.
func (r CreateIssueRequest) Validate() error {
	return check.All(
		check.Str(r.Title, "title").Required().V(),
	).Err()
}
