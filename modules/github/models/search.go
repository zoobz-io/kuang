package models

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
type CodeSearchResult struct {
	Results []SearchResult `json:"items"`
}

// Validate validates a CodeSearchResult.
func (c CodeSearchResult) Validate() error { return nil }
