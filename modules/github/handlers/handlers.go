// Package handlers defines the HTTP endpoints exposed by the GitHub module.
// Each endpoint resolves the API service from the sum registry at request
// time via sum.MustUse.
package handlers

import "github.com/zoobz-io/rocco"

// Endpoints returns all HTTP endpoints exposed by the GitHub module.
func Endpoints() []rocco.Endpoint {
	return []rocco.Endpoint{
		ListRepos, GetRepo,
		ListIssues, GetIssue, CreateIssue,
		ListPRs, GetPR, CreatePR,
		GetFile, CreateOrUpdateFile,
		SearchCode,
	}
}
