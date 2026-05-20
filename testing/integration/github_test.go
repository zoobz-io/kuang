//go:build testing

package integration

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/zoobz-io/kuang"
	"github.com/zoobz-io/kuang/github"
)

// ---------------------------------------------------------------------------
// Mock GitHub API
// ---------------------------------------------------------------------------

func newMockGitHub(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	// --- Repos ---
	mux.HandleFunc("GET /users/{owner}/repos", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, []map[string]any{
			{"name": "alpha", "full_name": "testowner/alpha", "description": "First repo", "default_branch": "main", "html_url": "https://github.com/testowner/alpha", "clone_url": "https://github.com/testowner/alpha.git", "language": "Go", "stargazers_count": 42, "open_issues_count": 3, "private": false, "fork": false},
			{"name": "beta", "full_name": "testowner/beta", "description": "", "default_branch": "develop", "html_url": "https://github.com/testowner/beta", "clone_url": "https://github.com/testowner/beta.git", "language": "Rust", "stargazers_count": 0, "open_issues_count": 0, "private": true, "fork": false},
			{"name": "gamma-fork", "full_name": "testowner/gamma-fork", "description": "A fork", "default_branch": "main", "html_url": "https://github.com/testowner/gamma-fork", "clone_url": "https://github.com/testowner/gamma-fork.git", "language": "Go", "stargazers_count": 1, "open_issues_count": 0, "private": false, "fork": true},
		})
	})

	mux.HandleFunc("GET /repos/{owner}/{repo}", func(w http.ResponseWriter, r *http.Request) {
		repo := r.PathValue("repo")
		switch repo {
		case "not-found":
			http.Error(w, `{"message":"Not Found"}`, http.StatusNotFound)
		case "server-error":
			http.Error(w, `{"message":"Internal Server Error"}`, http.StatusInternalServerError)
		default:
			writeJSON(w, map[string]any{
				"name": repo, "full_name": "testowner/" + repo, "description": "Repo " + repo,
				"default_branch": "main", "html_url": "https://github.com/testowner/" + repo,
				"clone_url": "https://github.com/testowner/" + repo + ".git",
				"language": "Go", "stargazers_count": 10, "open_issues_count": 2,
				"private": false, "fork": false,
			})
		}
	})

	// --- Issues ---
	mux.HandleFunc("GET /repos/{owner}/{repo}/issues", func(w http.ResponseWriter, r *http.Request) {
		repo := r.PathValue("repo")
		if repo == "empty-repo" {
			writeJSON(w, []any{})
			return
		}
		writeJSON(w, []map[string]any{
			{"number": 1, "title": "Bug: crash on startup", "body": "App crashes immediately", "state": "open", "html_url": "https://github.com/testowner/" + repo + "/issues/1", "labels": []map[string]string{{"name": "bug", "color": "d73a4a"}}},
			{"number": 2, "title": "Feature: add logging", "body": "Need structured logging", "state": "open", "html_url": "https://github.com/testowner/" + repo + "/issues/2", "labels": []map[string]string{{"name": "enhancement", "color": "a2eeef"}, {"name": "good first issue", "color": "7057ff"}}},
			{"number": 3, "title": "Docs: update README", "body": "", "state": "closed", "html_url": "https://github.com/testowner/" + repo + "/issues/3", "labels": []map[string]string{}},
		})
	})

	mux.HandleFunc("GET /repos/{owner}/{repo}/issues/{number}", func(w http.ResponseWriter, r *http.Request) {
		number := r.PathValue("number")
		if number == "999" {
			http.Error(w, `{"message":"Not Found"}`, http.StatusNotFound)
			return
		}
		writeJSON(w, map[string]any{
			"number": json.Number(number), "title": "Issue #" + number, "body": "Issue body",
			"state": "open", "html_url": "https://github.com/testowner/repo/issues/" + number,
			"labels": []map[string]string{{"name": "bug", "color": "d73a4a"}},
		})
	})

	mux.HandleFunc("POST /repos/{owner}/{repo}/issues", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, `{"message":"Bad Request"}`, http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		writeJSON(w, map[string]any{
			"number": 42, "title": body["title"], "body": body["body"],
			"state": "open", "html_url": "https://github.com/testowner/repo/issues/42",
			"labels": []any{},
		})
	})

	// --- Pull Requests ---
	mux.HandleFunc("GET /repos/{owner}/{repo}/pulls", func(w http.ResponseWriter, r *http.Request) {
		repo := r.PathValue("repo")
		if repo == "empty-repo" {
			writeJSON(w, []any{})
			return
		}
		writeJSON(w, []map[string]any{
			{"number": 10, "title": "Add feature X", "body": "Implements X", "state": "open", "html_url": "https://github.com/testowner/" + repo + "/pull/10", "draft": false, "merged": false, "head": map[string]string{"ref": "feature-x", "sha": "abc123"}, "base": map[string]string{"ref": "main", "sha": "def456"}},
			{"number": 11, "title": "WIP: refactor", "body": "Draft PR", "state": "open", "html_url": "https://github.com/testowner/" + repo + "/pull/11", "draft": true, "merged": false, "head": map[string]string{"ref": "refactor", "sha": "111222"}, "base": map[string]string{"ref": "main", "sha": "def456"}},
			{"number": 9, "title": "Fix bug Y", "body": "Merged fix", "state": "closed", "html_url": "https://github.com/testowner/" + repo + "/pull/9", "draft": false, "merged": true, "head": map[string]string{"ref": "fix-y", "sha": "aaa111"}, "base": map[string]string{"ref": "main", "sha": "bbb222"}},
		})
	})

	mux.HandleFunc("GET /repos/{owner}/{repo}/pulls/{number}", func(w http.ResponseWriter, r *http.Request) {
		number := r.PathValue("number")
		if number == "999" {
			http.Error(w, `{"message":"Not Found"}`, http.StatusNotFound)
			return
		}
		writeJSON(w, map[string]any{
			"number": json.Number(number), "title": "PR #" + number, "body": "PR body",
			"state": "open", "html_url": "https://github.com/testowner/repo/pull/" + number,
			"draft": false, "merged": false,
			"head": map[string]string{"ref": "feature", "sha": "abc123"},
			"base": map[string]string{"ref": "main", "sha": "def456"},
		})
	})

	mux.HandleFunc("POST /repos/{owner}/{repo}/pulls", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, `{"message":"Bad Request"}`, http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		writeJSON(w, map[string]any{
			"number": 50, "title": body["title"], "body": body["body"],
			"state": "open", "html_url": "https://github.com/testowner/repo/pull/50",
			"draft": false, "merged": false,
			"head": map[string]string{"ref": body["head"], "sha": "newsha1"},
			"base": map[string]string{"ref": body["base"], "sha": "newsha2"},
		})
	})

	// --- Content ---
	mux.HandleFunc("GET /repos/{owner}/{repo}/contents/{path...}", func(w http.ResponseWriter, r *http.Request) {
		filePath := r.PathValue("path")
		if filePath == "missing.txt" {
			http.Error(w, `{"message":"Not Found"}`, http.StatusNotFound)
			return
		}
		ref := r.URL.Query().Get("ref")
		sha := "sha256abc"
		if ref != "" {
			sha = "sha256-" + ref
		}
		writeJSON(w, map[string]any{
			"path": filePath, "content": "SGVsbG8gV29ybGQ=", "sha": sha,
		})
	})

	mux.HandleFunc("PUT /repos/{owner}/{repo}/contents/{path...}", func(w http.ResponseWriter, r *http.Request) {
		filePath := r.PathValue("path")
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		writeJSON(w, map[string]any{
			"path": filePath, "content": body["content"], "sha": "newsha789",
		})
	})

	// --- Search ---
	mux.HandleFunc("GET /search/code", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		if q == "" {
			writeJSON(w, map[string]any{"items": []any{}})
			return
		}
		writeJSON(w, map[string]any{
			"items": []map[string]any{
				{"name": "main.go", "path": "cmd/main.go", "html_url": "https://github.com/testowner/alpha/blob/main/cmd/main.go", "repository": map[string]any{"name": "alpha", "full_name": "testowner/alpha"}},
				{"name": "config.go", "path": "internal/config.go", "html_url": "https://github.com/testowner/beta/blob/main/internal/config.go", "repository": map[string]any{"name": "beta", "full_name": "testowner/beta"}},
			},
		})
	})

	// Catch-all for unmatched routes.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Read and discard body to prevent connection issues.
		_, _ = io.ReadAll(r.Body)
		http.Error(w, `{"message":"Not Found"}`, http.StatusNotFound)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

// ---------------------------------------------------------------------------
// Test environment setup
// ---------------------------------------------------------------------------

func setupGitHub(t *testing.T) (*kuang.TestEnv, *httptest.Server) {
	t.Helper()
	mock := newMockGitHub(t)
	env := kuang.NewTestEnv(t, github.Scopes(),
		github.Module(github.Config{
			APIURL: mock.URL,
			Owner:  "testowner",
		}),
	)
	return env, mock
}

// ---------------------------------------------------------------------------
// Scope filtering tests
// ---------------------------------------------------------------------------

func TestScopeFiltering(t *testing.T) {
	env, _ := setupGitHub(t)

	t.Run("full_access_sees_all_tools", func(t *testing.T) {
		agent := issueAgent(t, env, "full", github.Scopes())
		names := mcpToolNames(t, agent)
		t.Logf("full agent tools: %v", names)

		// Should see all 11 endpoints.
		if len(names) != 11 {
			t.Errorf("expected 11 tools, got %d: %v", len(names), names)
		}
	})

	t.Run("read_only_sees_read_tools", func(t *testing.T) {
		readScopes := []string{
			github.ScopeReposRead, github.ScopeIssuesRead,
			github.ScopePullsRead, github.ScopeContentRead,
			github.ScopeSearchRead,
		}
		agent := issueAgent(t, env, "reader", readScopes)
		names := mcpToolNames(t, agent)
		t.Logf("reader tools: %v", names)

		// 2 repo reads + 2 issue reads + 2 PR reads + 2 content reads + 1 search = 9
		// Wait - getFile has path..., only 1 GET for content.
		// listRepos, getRepo, listIssues, getIssue, listPRs, getPR, getFile, searchCode = 8
		if len(names) != 8 {
			t.Errorf("expected 8 read tools, got %d: %v", len(names), names)
		}

		// Should NOT see any write tools.
		for _, n := range names {
			lower := strings.ToLower(n)
			if strings.Contains(lower, "create") || strings.Contains(lower, "update") {
				t.Errorf("read-only agent should not see write tool %q", n)
			}
		}
	})

	t.Run("repos_only_sees_repo_tools", func(t *testing.T) {
		agent := issueAgent(t, env, "repos-only", []string{github.ScopeReposRead})
		names := mcpToolNames(t, agent)
		t.Logf("repos-only tools: %v", names)

		// Only listRepos and getRepo.
		if len(names) != 2 {
			t.Errorf("expected 2 tools, got %d: %v", len(names), names)
		}
	})

	t.Run("no_scopes_sees_no_tools", func(t *testing.T) {
		agent := issueAgent(t, env, "empty", nil)
		names := mcpToolNames(t, agent)
		t.Logf("empty agent tools: %v", names)

		if len(names) != 0 {
			t.Errorf("expected 0 tools, got %d: %v", len(names), names)
		}
	})

	t.Run("write_only_sees_write_tools", func(t *testing.T) {
		writeScopes := []string{
			github.ScopeIssuesWrite, github.ScopePullsWrite, github.ScopeContentWrite,
		}
		agent := issueAgent(t, env, "writer", writeScopes)
		names := mcpToolNames(t, agent)
		t.Logf("writer tools: %v", names)

		// createIssue, createPR, createOrUpdateFile = 3
		if len(names) != 3 {
			t.Errorf("expected 3 write tools, got %d: %v", len(names), names)
		}
	})
}

// ---------------------------------------------------------------------------
// Repo operation tests
// ---------------------------------------------------------------------------

func TestRepos(t *testing.T) {
	env, _ := setupGitHub(t)
	agent := fullAccessAgent(t, env, github.Scopes())
	names := mcpToolNames(t, agent)

	t.Run("list_repos", func(t *testing.T) {
		tool := toolByName(t, names, "listRepos")
		text, isErr := mcpToolCall(t, agent, tool, `{}`, 100)
		if isErr {
			t.Fatalf("list repos failed: %s", text)
		}
		if !strings.Contains(text, "alpha") {
			t.Errorf("expected 'alpha' in response, got: %s", text)
		}
		if !strings.Contains(text, "beta") {
			t.Errorf("expected 'beta' in response")
		}
		if !strings.Contains(text, "gamma-fork") {
			t.Errorf("expected 'gamma-fork' in response")
		}
	})

	t.Run("get_repo", func(t *testing.T) {
		tool := toolByName(t, names, "getRepo")
		text, isErr := mcpToolCall(t, agent, tool, `{"name":"alpha"}`, 101)
		if isErr {
			t.Fatalf("get repo failed: %s", text)
		}
		if !strings.Contains(text, "alpha") {
			t.Errorf("expected 'alpha' in response, got: %s", text)
		}
	})

	t.Run("get_repo_not_found", func(t *testing.T) {
		tool := toolByName(t, names, "getRepo")
		_, isErr := mcpToolCall(t, agent, tool, `{"name":"not-found"}`, 102)
		if !isErr {
			t.Error("expected error for not-found repo")
		}
	})

	t.Run("get_repo_server_error", func(t *testing.T) {
		tool := toolByName(t, names, "getRepo")
		_, isErr := mcpToolCall(t, agent, tool, `{"name":"server-error"}`, 103)
		if !isErr {
			t.Error("expected error for server-error repo")
		}
	})
}

// ---------------------------------------------------------------------------
// Issue operation tests
// ---------------------------------------------------------------------------

func TestIssues(t *testing.T) {
	env, _ := setupGitHub(t)
	agent := fullAccessAgent(t, env, github.Scopes())
	names := mcpToolNames(t, agent)

	t.Run("list_issues", func(t *testing.T) {
		tool := toolByName(t, names, "listIssues")
		text, isErr := mcpToolCall(t, agent, tool, `{"repo":"alpha"}`, 200)
		if isErr {
			t.Fatalf("list issues failed: %s", text)
		}
		if !strings.Contains(text, "crash on startup") {
			t.Errorf("expected 'crash on startup' in response")
		}
		if !strings.Contains(text, "bug") {
			t.Errorf("expected label 'bug' in response")
		}
	})

	t.Run("list_issues_empty", func(t *testing.T) {
		tool := toolByName(t, names, "listIssues")
		text, isErr := mcpToolCall(t, agent, tool, `{"repo":"empty-repo"}`, 201)
		if isErr {
			t.Fatalf("list issues (empty) failed: %s", text)
		}
		// Should succeed with an empty list.
		if !strings.Contains(text, "[]") && !strings.Contains(text, "issues") {
			t.Logf("empty list response: %s", text)
		}
	})

	t.Run("get_issue", func(t *testing.T) {
		tool := toolByName(t, names, "getIssue")
		text, isErr := mcpToolCall(t, agent, tool, `{"repo":"alpha","number":"1"}`, 202)
		if isErr {
			t.Fatalf("get issue failed: %s", text)
		}
		if !strings.Contains(text, "Issue #1") {
			t.Errorf("expected 'Issue #1' in response, got: %s", text)
		}
	})

	t.Run("get_issue_not_found", func(t *testing.T) {
		tool := toolByName(t, names, "getIssue")
		_, isErr := mcpToolCall(t, agent, tool, `{"repo":"alpha","number":"999"}`, 203)
		if !isErr {
			t.Error("expected error for issue 999")
		}
	})

	t.Run("create_issue", func(t *testing.T) {
		tool := toolByName(t, names, "createIssue")
		text, isErr := mcpToolCall(t, agent, tool, `{"repo":"alpha","body":{"title":"New bug","body":"Description here"}}`, 204)
		if isErr {
			t.Fatalf("create issue failed: %s", text)
		}
		if !strings.Contains(text, "New bug") {
			t.Errorf("expected 'New bug' in response, got: %s", text)
		}
	})
}

// ---------------------------------------------------------------------------
// Pull request operation tests
// ---------------------------------------------------------------------------

func TestPullRequests(t *testing.T) {
	env, _ := setupGitHub(t)
	agent := fullAccessAgent(t, env, github.Scopes())
	names := mcpToolNames(t, agent)

	t.Run("list_prs", func(t *testing.T) {
		tool := toolByName(t, names, "listPRs")
		text, isErr := mcpToolCall(t, agent, tool, `{"repo":"alpha"}`, 300)
		if isErr {
			t.Fatalf("list PRs failed: %s", text)
		}
		if !strings.Contains(text, "Add feature X") {
			t.Errorf("expected 'Add feature X' in response")
		}
		// Verify draft and merged states are present.
		if !strings.Contains(text, "draft") {
			t.Errorf("expected draft field in response")
		}
	})

	t.Run("list_prs_empty", func(t *testing.T) {
		tool := toolByName(t, names, "listPRs")
		text, isErr := mcpToolCall(t, agent, tool, `{"repo":"empty-repo"}`, 301)
		if isErr {
			t.Fatalf("list PRs (empty) failed: %s", text)
		}
		t.Logf("empty PR list: %s", text)
	})

	t.Run("get_pr", func(t *testing.T) {
		tool := toolByName(t, names, "getPR")
		text, isErr := mcpToolCall(t, agent, tool, `{"repo":"alpha","number":"10"}`, 302)
		if isErr {
			t.Fatalf("get PR failed: %s", text)
		}
		if !strings.Contains(text, "PR #10") {
			t.Errorf("expected 'PR #10' in response, got: %s", text)
		}
		// Verify branch ref fields.
		if !strings.Contains(text, "feature") || !strings.Contains(text, "main") {
			t.Errorf("expected head/base refs in response")
		}
	})

	t.Run("get_pr_not_found", func(t *testing.T) {
		tool := toolByName(t, names, "getPR")
		_, isErr := mcpToolCall(t, agent, tool, `{"repo":"alpha","number":"999"}`, 303)
		if !isErr {
			t.Error("expected error for PR 999")
		}
	})

	t.Run("create_pr", func(t *testing.T) {
		tool := toolByName(t, names, "createPR")
		args := `{"repo":"alpha","body":{"title":"New PR","body":"PR description","head":"feature-branch","base":"main"}}`
		text, isErr := mcpToolCall(t, agent, tool, args, 304)
		if isErr {
			t.Fatalf("create PR failed: %s", text)
		}
		if !strings.Contains(text, "New PR") {
			t.Errorf("expected 'New PR' in response, got: %s", text)
		}
		if !strings.Contains(text, "feature-branch") {
			t.Errorf("expected 'feature-branch' in response")
		}
	})
}

// ---------------------------------------------------------------------------
// Content operation tests
// ---------------------------------------------------------------------------

func TestContent(t *testing.T) {
	env, _ := setupGitHub(t)
	agent := fullAccessAgent(t, env, github.Scopes())
	names := mcpToolNames(t, agent)

	t.Run("get_file", func(t *testing.T) {
		tool := toolByName(t, names, "getFile")
		text, isErr := mcpToolCall(t, agent, tool, `{"repo":"alpha","path...":"src/main.go"}`, 400)
		if isErr {
			t.Fatalf("get file failed: %s", text)
		}
		if !strings.Contains(text, "src/main.go") {
			t.Errorf("expected file path in response, got: %s", text)
		}
		// Base64 content should be present.
		if !strings.Contains(text, "SGVsbG8gV29ybGQ=") {
			t.Errorf("expected base64 content in response")
		}
	})

	t.Run("get_file_with_ref", func(t *testing.T) {
		tool := toolByName(t, names, "getFile")
		text, isErr := mcpToolCall(t, agent, tool, `{"repo":"alpha","path...":"README.md","ref":"v1.0"}`, 401)
		if isErr {
			t.Fatalf("get file with ref failed: %s", text)
		}
		if !strings.Contains(text, "sha256-v1.0") {
			t.Errorf("expected ref-specific SHA in response, got: %s", text)
		}
	})

	t.Run("get_file_not_found", func(t *testing.T) {
		tool := toolByName(t, names, "getFile")
		_, isErr := mcpToolCall(t, agent, tool, `{"repo":"alpha","path...":"missing.txt"}`, 402)
		if !isErr {
			t.Error("expected error for missing file")
		}
	})

	t.Run("create_or_update_file", func(t *testing.T) {
		tool := toolByName(t, names, "createOrUpdateFile")
		args := `{"repo":"alpha","path...":"new-file.md","body":{"message":"Add file","content":"IyBIZWxsbw=="}}`
		text, isErr := mcpToolCall(t, agent, tool, args, 403)
		if isErr {
			t.Fatalf("create file failed: %s", text)
		}
		if !strings.Contains(text, "new-file.md") {
			t.Errorf("expected file path in response, got: %s", text)
		}
	})

	t.Run("update_file_with_sha", func(t *testing.T) {
		tool := toolByName(t, names, "createOrUpdateFile")
		args := `{"repo":"alpha","path...":"existing.md","body":{"message":"Update file","content":"dXBkYXRlZA==","sha":"oldsha123"}}`
		text, isErr := mcpToolCall(t, agent, tool, args, 404)
		if isErr {
			t.Fatalf("update file failed: %s", text)
		}
		if !strings.Contains(text, "existing.md") {
			t.Errorf("expected file path in response, got: %s", text)
		}
	})
}

// ---------------------------------------------------------------------------
// Search operation tests
// ---------------------------------------------------------------------------

func TestSearch(t *testing.T) {
	env, _ := setupGitHub(t)
	agent := fullAccessAgent(t, env, github.Scopes())
	names := mcpToolNames(t, agent)

	t.Run("search_code", func(t *testing.T) {
		tool := toolByName(t, names, "searchCode")
		text, isErr := mcpToolCall(t, agent, tool, `{"query":"main"}`, 500)
		if isErr {
			t.Fatalf("search failed: %s", text)
		}
		if !strings.Contains(text, "main.go") {
			t.Errorf("expected 'main.go' in results, got: %s", text)
		}
		if !strings.Contains(text, "config.go") {
			t.Errorf("expected 'config.go' in results")
		}
	})

	t.Run("search_code_empty_query", func(t *testing.T) {
		tool := toolByName(t, names, "searchCode")
		text, isErr := mcpToolCall(t, agent, tool, `{"query":""}`, 501)
		if isErr {
			t.Fatalf("empty search failed: %s", text)
		}
		t.Logf("empty search result: %s", text)
	})
}

// ---------------------------------------------------------------------------
// Cross-cutting: read-only agent can read but not write
// ---------------------------------------------------------------------------

func TestReadOnlyAgentOperations(t *testing.T) {
	env, _ := setupGitHub(t)
	readScopes := []string{
		github.ScopeReposRead, github.ScopeIssuesRead,
		github.ScopePullsRead, github.ScopeContentRead,
		github.ScopeSearchRead,
	}
	agent := issueAgent(t, env, "read-ops", readScopes)
	names := mcpToolNames(t, agent)

	t.Run("can_list_repos", func(t *testing.T) {
		tool := toolByName(t, names, "listRepos")
		text, isErr := mcpToolCall(t, agent, tool, `{}`, 600)
		if isErr {
			t.Fatalf("read agent should list repos: %s", text)
		}
		if !strings.Contains(text, "alpha") {
			t.Errorf("expected 'alpha' in response")
		}
	})

	t.Run("cannot_see_create_issue", func(t *testing.T) {
		if hasToolMatching(names, "createIssue") {
			t.Error("read-only agent should not see createIssue tool")
		}
	})

	t.Run("cannot_see_create_pr", func(t *testing.T) {
		if hasToolMatching(names, "createPR") {
			t.Error("read-only agent should not see createPR tool")
		}
	})

	t.Run("cannot_see_create_or_update_file", func(t *testing.T) {
		if hasToolMatching(names, "createOrUpdateFile") {
			t.Error("read-only agent should not see createOrUpdateFile tool")
		}
	})
}
