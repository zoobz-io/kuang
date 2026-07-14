package models

import "github.com/zoobz-io/check"

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
	Message string `json:"message"`
	Content string `json:"content"`
	SHA     string `json:"sha"`
}

// Validate validates a CreateOrUpdateFileRequest.
func (r CreateOrUpdateFileRequest) Validate() error {
	return check.All(
		check.Str(r.Message, "message").Required().V(),
		check.Str(r.Content, "content").Required().V(),
	).Err()
}
