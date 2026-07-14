// Package models holds the API data models for the GitHub module.
package models

import "github.com/zoobz-io/check"

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
