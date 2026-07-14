package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/zoobzio/kuang/internal/httpc"
)

// testService creates an API service backed by the given test server.
func testService(ts *httptest.Server) *API {
	return &API{
		client: httpc.New(httpc.WithBaseURL(ts.URL)),
		owner:  "octocat",
	}
}

// jsonHandler returns an http.HandlerFunc that responds with the given value as JSON.
func jsonHandler(v any) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(v)
	}
}

// newTestServer registers the given routes and returns the test server.
func newTestServer(routes map[string]http.HandlerFunc) *httptest.Server {
	mux := http.NewServeMux()
	for pattern, handler := range routes {
		mux.HandleFunc(pattern, handler)
	}
	return httptest.NewServer(mux)
}

// --- Construction ---

// clientBaseURL reads the unexported baseURL field of an httpc.Client.
func clientBaseURL(c *httpc.Client) string {
	return reflect.ValueOf(c).Elem().FieldByName("baseURL").String()
}

// clientHeader reads a header value from the unexported headers map of an httpc.Client.
func clientHeader(c *httpc.Client, key string) string {
	headers := reflect.ValueOf(c).Elem().FieldByName("headers")
	v := headers.MapIndex(reflect.ValueOf(key))
	if !v.IsValid() {
		return ""
	}
	return v.String()
}

func TestNewAPISetsBaseURL(t *testing.T) {
	api := NewAPI(Config{Owner: "octocat", APIURL: "https://api.github.com"})
	if got := clientBaseURL(api.client); got != "https://api.github.com" {
		t.Errorf("baseURL = %q, want https://api.github.com", got)
	}
	if api.owner != "octocat" {
		t.Errorf("owner = %q, want octocat", api.owner)
	}
}

func TestNewAPIExplicitBaseURL(t *testing.T) {
	api := NewAPI(Config{Owner: "octocat", APIURL: "https://ghe.example.com/api/v3"})
	if got := clientBaseURL(api.client); got != "https://ghe.example.com/api/v3" {
		t.Errorf("baseURL = %q, want explicit URL", got)
	}
}

func TestNewAPISetsGitHubHeaders(t *testing.T) {
	api := NewAPI(Config{Owner: "octocat", APIURL: "https://api.github.com"})
	if got := clientHeader(api.client, "Accept"); got != "application/vnd.github+json" {
		t.Errorf("Accept header = %q, want application/vnd.github+json", got)
	}
	if got := clientHeader(api.client, "X-GitHub-Api-Version"); got != "2022-11-28" {
		t.Errorf("X-GitHub-Api-Version header = %q, want 2022-11-28", got)
	}
}

// --- Repos ---

func TestListRepos(t *testing.T) {
	var gotPath string
	ts := newTestServer(map[string]http.HandlerFunc{
		"GET /users/octocat/repos": func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"name": "hello"},
				{"name": "world"},
			})
		},
	})
	defer ts.Close()

	svc := testService(ts)
	list, err := svc.ListRepos(context.Background())
	if err != nil {
		t.Fatalf("ListRepos: %v", err)
	}
	if gotPath != "/users/octocat/repos" {
		t.Errorf("path = %q, want /users/octocat/repos", gotPath)
	}
	if len(list.Repos) != 2 {
		t.Fatalf("repos = %d, want 2", len(list.Repos))
	}
	if list.Repos[0].Name != "hello" {
		t.Errorf("repos[0].name = %q, want hello", list.Repos[0].Name)
	}
}

func TestGetRepo(t *testing.T) {
	ts := newTestServer(map[string]http.HandlerFunc{
		"GET /repos/octocat/hello": jsonHandler(map[string]any{
			"name":      "hello",
			"full_name": "octocat/hello",
		}),
	})
	defer ts.Close()

	svc := testService(ts)
	repo, err := svc.GetRepo(context.Background(), "hello")
	if err != nil {
		t.Fatalf("GetRepo: %v", err)
	}
	if repo.Name != "hello" {
		t.Errorf("name = %q, want hello", repo.Name)
	}
	if repo.FullName != "octocat/hello" {
		t.Errorf("full_name = %q, want octocat/hello", repo.FullName)
	}
}

// --- Issues ---

func TestListIssues(t *testing.T) {
	ts := newTestServer(map[string]http.HandlerFunc{
		"GET /repos/octocat/hello/issues": jsonHandler([]map[string]any{
			{"title": "first", "number": 1},
			{"title": "second", "number": 2},
		}),
	})
	defer ts.Close()

	svc := testService(ts)
	list, err := svc.ListIssues(context.Background(), "hello")
	if err != nil {
		t.Fatalf("ListIssues: %v", err)
	}
	if len(list.Issues) != 2 {
		t.Fatalf("issues = %d, want 2", len(list.Issues))
	}
	if list.Issues[0].Title != "first" {
		t.Errorf("issues[0].title = %q, want first", list.Issues[0].Title)
	}
}

func TestGetIssue(t *testing.T) {
	var gotPath string
	ts := newTestServer(map[string]http.HandlerFunc{
		"GET /repos/octocat/hello/issues/7": func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"title": "a bug", "number": 7})
		},
	})
	defer ts.Close()

	svc := testService(ts)
	issue, err := svc.GetIssue(context.Background(), "hello", 7)
	if err != nil {
		t.Fatalf("GetIssue: %v", err)
	}
	if gotPath != "/repos/octocat/hello/issues/7" {
		t.Errorf("path = %q", gotPath)
	}
	if issue.Number != 7 {
		t.Errorf("number = %d, want 7", issue.Number)
	}
}

func TestCreateIssue(t *testing.T) {
	var gotBody map[string]string
	ts := newTestServer(map[string]http.HandlerFunc{
		"POST /repos/octocat/hello/issues": func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewDecoder(r.Body).Decode(&gotBody)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"title": "new issue", "number": 42})
		},
	})
	defer ts.Close()

	svc := testService(ts)
	issue, err := svc.CreateIssue(context.Background(), "hello", "new issue", "the body")
	if err != nil {
		t.Fatalf("CreateIssue: %v", err)
	}
	if gotBody["title"] != "new issue" {
		t.Errorf("body title = %q, want new issue", gotBody["title"])
	}
	if gotBody["body"] != "the body" {
		t.Errorf("body body = %q, want the body", gotBody["body"])
	}
	if issue.Number != 42 {
		t.Errorf("number = %d, want 42", issue.Number)
	}
}

// --- Pull Requests ---

func TestListPRs(t *testing.T) {
	ts := newTestServer(map[string]http.HandlerFunc{
		"GET /repos/octocat/hello/pulls": jsonHandler([]map[string]any{
			{"title": "pr one", "number": 1},
		}),
	})
	defer ts.Close()

	svc := testService(ts)
	list, err := svc.ListPRs(context.Background(), "hello")
	if err != nil {
		t.Fatalf("ListPRs: %v", err)
	}
	if len(list.PullRequests) != 1 {
		t.Fatalf("prs = %d, want 1", len(list.PullRequests))
	}
	if list.PullRequests[0].Title != "pr one" {
		t.Errorf("prs[0].title = %q, want pr one", list.PullRequests[0].Title)
	}
}

func TestGetPR(t *testing.T) {
	var gotPath string
	ts := newTestServer(map[string]http.HandlerFunc{
		"GET /repos/octocat/hello/pulls/3": func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"title": "a pr", "number": 3})
		},
	})
	defer ts.Close()

	svc := testService(ts)
	pr, err := svc.GetPR(context.Background(), "hello", 3)
	if err != nil {
		t.Fatalf("GetPR: %v", err)
	}
	if gotPath != "/repos/octocat/hello/pulls/3" {
		t.Errorf("path = %q", gotPath)
	}
	if pr.Number != 3 {
		t.Errorf("number = %d, want 3", pr.Number)
	}
}

func TestCreatePR(t *testing.T) {
	var gotBody map[string]string
	ts := newTestServer(map[string]http.HandlerFunc{
		"POST /repos/octocat/hello/pulls": func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewDecoder(r.Body).Decode(&gotBody)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"title": "feat", "number": 9})
		},
	})
	defer ts.Close()

	svc := testService(ts)
	pr, err := svc.CreatePR(context.Background(), "hello", "feat", "desc", "feature", "main")
	if err != nil {
		t.Fatalf("CreatePR: %v", err)
	}
	if gotBody["title"] != "feat" {
		t.Errorf("body title = %q, want feat", gotBody["title"])
	}
	if gotBody["head"] != "feature" {
		t.Errorf("body head = %q, want feature", gotBody["head"])
	}
	if gotBody["base"] != "main" {
		t.Errorf("body base = %q, want main", gotBody["base"])
	}
	if pr.Number != 9 {
		t.Errorf("number = %d, want 9", pr.Number)
	}
}

// --- Content ---

func TestGetFile(t *testing.T) {
	var gotRef string
	ts := newTestServer(map[string]http.HandlerFunc{
		"GET /repos/octocat/hello/contents/README.md": func(w http.ResponseWriter, r *http.Request) {
			gotRef = r.URL.Query().Get("ref")
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"path":    "README.md",
				"content": "aGVsbG8=",
				"sha":     "abc123",
			})
		},
	})
	defer ts.Close()

	svc := testService(ts)
	fc, err := svc.GetFile(context.Background(), "hello", "README.md", "")
	if err != nil {
		t.Fatalf("GetFile: %v", err)
	}
	if gotRef != "" {
		t.Errorf("ref = %q, want empty", gotRef)
	}
	if fc.Path != "README.md" {
		t.Errorf("path = %q, want README.md", fc.Path)
	}
	if fc.SHA != "abc123" {
		t.Errorf("sha = %q, want abc123", fc.SHA)
	}
}

func TestGetFileWithRef(t *testing.T) {
	var gotRef string
	ts := newTestServer(map[string]http.HandlerFunc{
		"GET /repos/octocat/hello/contents/README.md": func(w http.ResponseWriter, r *http.Request) {
			gotRef = r.URL.Query().Get("ref")
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"path": "README.md"})
		},
	})
	defer ts.Close()

	svc := testService(ts)
	if _, err := svc.GetFile(context.Background(), "hello", "README.md", "develop"); err != nil {
		t.Fatalf("GetFile with ref: %v", err)
	}
	if gotRef != "develop" {
		t.Errorf("ref = %q, want develop", gotRef)
	}
}

func TestCreateOrUpdateFileCreate(t *testing.T) {
	var gotBody map[string]string
	ts := newTestServer(map[string]http.HandlerFunc{
		"PUT /repos/octocat/hello/contents/docs/x.md": func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewDecoder(r.Body).Decode(&gotBody)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"path": "docs/x.md", "sha": "newsha"})
		},
	})
	defer ts.Close()

	svc := testService(ts)
	fc, err := svc.CreateOrUpdateFile(context.Background(), "hello", "docs/x.md", "add docs", "Y29udGVudA==", "")
	if err != nil {
		t.Fatalf("CreateOrUpdateFile: %v", err)
	}
	if gotBody["message"] != "add docs" {
		t.Errorf("message = %q, want add docs", gotBody["message"])
	}
	if gotBody["content"] != "Y29udGVudA==" {
		t.Errorf("content = %q", gotBody["content"])
	}
	if _, ok := gotBody["sha"]; ok {
		t.Errorf("sha should be omitted for create, got %q", gotBody["sha"])
	}
	if fc.SHA != "newsha" {
		t.Errorf("sha = %q, want newsha", fc.SHA)
	}
}

func TestCreateOrUpdateFileUpdate(t *testing.T) {
	var gotBody map[string]string
	ts := newTestServer(map[string]http.HandlerFunc{
		"PUT /repos/octocat/hello/contents/docs/x.md": func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewDecoder(r.Body).Decode(&gotBody)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"path": "docs/x.md", "sha": "updated"})
		},
	})
	defer ts.Close()

	svc := testService(ts)
	if _, err := svc.CreateOrUpdateFile(context.Background(), "hello", "docs/x.md", "update", "bmV3", "oldsha"); err != nil {
		t.Fatalf("CreateOrUpdateFile update: %v", err)
	}
	if gotBody["sha"] != "oldsha" {
		t.Errorf("sha = %q, want oldsha (update path)", gotBody["sha"])
	}
}

// --- Search ---

func TestSearchCode(t *testing.T) {
	var gotQuery string
	ts := newTestServer(map[string]http.HandlerFunc{
		"GET /search/code": func(w http.ResponseWriter, r *http.Request) {
			gotQuery = r.URL.Query().Get("q")
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"items": []map[string]any{
					{"name": "main.go", "path": "cmd/main.go"},
				},
			})
		},
	})
	defer ts.Close()

	svc := testService(ts)
	result, err := svc.SearchCode(context.Background(), "addClass")
	if err != nil {
		t.Fatalf("SearchCode: %v", err)
	}
	if gotQuery != "addClass" {
		t.Errorf("query = %q, want addClass", gotQuery)
	}
	if len(result.Results) != 1 {
		t.Fatalf("results = %d, want 1", len(result.Results))
	}
	if result.Results[0].Name != "main.go" {
		t.Errorf("results[0].name = %q, want main.go", result.Results[0].Name)
	}
}
