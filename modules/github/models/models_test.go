package models

import "testing"

func TestCreateIssueRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateIssueRequest
		wantErr bool
	}{
		{"valid", CreateIssueRequest{Title: "bug"}, false},
		{"valid with body", CreateIssueRequest{Title: "bug", Body: "details"}, false},
		{"missing title", CreateIssueRequest{Body: "details"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.req.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreatePRRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		req     CreatePRRequest
		wantErr bool
	}{
		{"valid", CreatePRRequest{Title: "feat", Head: "feature", Base: "main"}, false},
		{"missing title", CreatePRRequest{Head: "feature", Base: "main"}, true},
		{"missing head", CreatePRRequest{Title: "feat", Base: "main"}, true},
		{"missing base", CreatePRRequest{Title: "feat", Head: "feature"}, true},
		{"empty", CreatePRRequest{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.req.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateOrUpdateFileRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateOrUpdateFileRequest
		wantErr bool
	}{
		{"valid", CreateOrUpdateFileRequest{Message: "add", Content: "aGk="}, false},
		{"missing message", CreateOrUpdateFileRequest{Content: "aGk="}, true},
		{"missing content", CreateOrUpdateFileRequest{Message: "add"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.req.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRequiredFieldModelsValidate(t *testing.T) {
	tests := []struct {
		name    string
		v       interface{ Validate() error }
		wantErr bool
	}{
		{"issue valid", Issue{Title: "t"}, false},
		{"issue missing title", Issue{}, true},
		{"pull request valid", PullRequest{Title: "t"}, false},
		{"pull request missing title", PullRequest{}, true},
		{"repo valid", Repo{Name: "n"}, false},
		{"repo missing name", Repo{}, true},
		{"file content valid", FileContent{Path: "p"}, false},
		{"file content missing path", FileContent{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.v.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResponseModelsValidateNil(t *testing.T) {
	models := []interface{ Validate() error }{
		RepoList{},
		IssueList{},
		PRList{},
		SearchResult{},
		CodeSearchResult{},
	}
	for _, m := range models {
		if err := m.Validate(); err != nil {
			t.Errorf("%T.Validate() = %v, want nil", m, err)
		}
	}
}
