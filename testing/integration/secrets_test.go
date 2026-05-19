//go:build testing

package integration

import (
	"context"
	"strings"
	"testing"

	"github.com/zoobz-io/kuang"
	"github.com/zoobz-io/kuang/secrets"
)

// ---------------------------------------------------------------------------
// Stub secret provider
// ---------------------------------------------------------------------------

type memProvider struct {
	store map[string]string
}

func (m *memProvider) Get(_ context.Context, key string) (string, error) {
	if v, ok := m.store[key]; ok {
		return v, nil
	}
	return "", context.Canceled
}

// ---------------------------------------------------------------------------
// Setup
// ---------------------------------------------------------------------------

func setupSecrets(t *testing.T) (*kuang.TestEnv, secrets.Config) {
	t.Helper()
	cfg := secrets.Config{
		Provider: &memProvider{store: map[string]string{
			"db/password":      "hunter2",
			"db/host":          "localhost:5432",
			"api/github-token": "ghp_abc123",
		}},
		Scopes: map[string][]string{
			"secrets-db":  {"db/password", "db/host"},
			"secrets-api": {"api/github-token"},
		},
	}
	allScopes := secrets.Scopes(cfg)
	env := kuang.NewTestEnv(t, allScopes, secrets.Module(cfg))
	return env, cfg
}

// ---------------------------------------------------------------------------
// Scope filtering
// ---------------------------------------------------------------------------

func TestSecretsToolVisibility(t *testing.T) {
	env, cfg := setupSecrets(t)

	t.Run("agent_with_any_secret_scope_sees_endpoint", func(t *testing.T) {
		agent := issueAgent(t, env, "db-agent", []string{"secrets-db"})
		names := mcpToolNames(t, agent)
		if !hasToolMatching(names, "getSecret") {
			t.Errorf("expected getSecret tool, got: %v", names)
		}
	})

	t.Run("agent_with_all_secret_scopes_sees_endpoint", func(t *testing.T) {
		agent := fullAccessAgent(t, env, secrets.Scopes(cfg))
		names := mcpToolNames(t, agent)
		if !hasToolMatching(names, "getSecret") {
			t.Errorf("expected getSecret tool, got: %v", names)
		}
	})

	t.Run("agent_with_no_secret_scopes_sees_nothing", func(t *testing.T) {
		agent := issueAgent(t, env, "no-secrets", nil)
		names := mcpToolNames(t, agent)
		if hasToolMatching(names, "getSecret") {
			t.Error("agent without secret scopes should not see getSecret")
		}
	})
}

// ---------------------------------------------------------------------------
// Key-level access control
// ---------------------------------------------------------------------------

func TestSecretsKeyAccess(t *testing.T) {
	env, _ := setupSecrets(t)

	t.Run("db_agent_can_read_db_password", func(t *testing.T) {
		agent := issueAgent(t, env, "db-read", []string{"secrets-db"})
		text, isErr := mcpToolCall(t, agent, "getSecret", `{"key...":"db/password"}`, 100)
		if isErr {
			t.Fatalf("expected success, got: %s", text)
		}
		if !strings.Contains(text, "hunter2") {
			t.Errorf("expected secret value in response, got: %s", text)
		}
	})

	t.Run("db_agent_can_read_db_host", func(t *testing.T) {
		agent := issueAgent(t, env, "db-host", []string{"secrets-db"})
		text, isErr := mcpToolCall(t, agent, "getSecret", `{"key...":"db/host"}`, 101)
		if isErr {
			t.Fatalf("expected success, got: %s", text)
		}
		if !strings.Contains(text, "localhost:5432") {
			t.Errorf("expected host in response, got: %s", text)
		}
	})

	t.Run("db_agent_denied_api_secret", func(t *testing.T) {
		agent := issueAgent(t, env, "db-denied", []string{"secrets-db"})
		_, isErr := mcpToolCall(t, agent, "getSecret", `{"key...":"api/github-token"}`, 102)
		if !isErr {
			t.Error("db agent should not access api secret")
		}
	})

	t.Run("api_agent_can_read_api_token", func(t *testing.T) {
		agent := issueAgent(t, env, "api-read", []string{"secrets-api"})
		text, isErr := mcpToolCall(t, agent, "getSecret", `{"key...":"api/github-token"}`, 103)
		if isErr {
			t.Fatalf("expected success, got: %s", text)
		}
		if !strings.Contains(text, "ghp_abc123") {
			t.Errorf("expected token in response, got: %s", text)
		}
	})

	t.Run("api_agent_denied_db_secret", func(t *testing.T) {
		agent := issueAgent(t, env, "api-denied", []string{"secrets-api"})
		_, isErr := mcpToolCall(t, agent, "getSecret", `{"key...":"db/password"}`, 104)
		if !isErr {
			t.Error("api agent should not access db secret")
		}
	})

	t.Run("full_agent_reads_all", func(t *testing.T) {
		agent := issueAgent(t, env, "full-secrets", []string{"secrets-db", "secrets-api"})
		for _, key := range []string{"db/password", "db/host", "api/github-token"} {
			text, isErr := mcpToolCall(t, agent, "getSecret", `{"key...":"`+key+`"}`, 105)
			if isErr {
				t.Errorf("full agent denied %q: %s", key, text)
			}
		}
	})

	t.Run("unmapped_key_denied", func(t *testing.T) {
		agent := issueAgent(t, env, "unmapped", []string{"secrets-db", "secrets-api"})
		_, isErr := mcpToolCall(t, agent, "getSecret", `{"key...":"nonexistent/key"}`, 106)
		if !isErr {
			t.Error("expected denial for unmapped key")
		}
	})
}
