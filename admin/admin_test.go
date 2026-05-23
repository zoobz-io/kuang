//go:build testing

package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/zoobz-io/kuang/internal/creds"
	"github.com/zoobz-io/sum"
)

const testAPIKey = "test-key-12345"

func setupTestHandler(t *testing.T) http.Handler {
	t.Helper()
	sum.Reset()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := creds.Open(dbPath)
	if err != nil {
		t.Fatalf("open test store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return NewHandler(store, testAPIKey)
}

func authHeader() string {
	return "Bearer " + testAPIKey
}

// --- Auth middleware tests ---

func TestAuthenticateRejectsNoHeader(t *testing.T) {
	handler := setupTestHandler(t)
	req := httptest.NewRequest("GET", "/credentials/agent-1", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestAuthenticateRejectsWrongKey(t *testing.T) {
	handler := setupTestHandler(t)
	req := httptest.NewRequest("GET", "/credentials/agent-1", nil)
	req.Header.Set("Authorization", "Bearer wrong-key")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestAuthenticateAcceptsValidKey(t *testing.T) {
	handler := setupTestHandler(t)
	req := httptest.NewRequest("GET", "/credentials/agent-1", nil)
	req.Header.Set("Authorization", authHeader())
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

// --- AdminIdentity tests ---

func TestAdminIdentity(t *testing.T) {
	id := &AdminIdentity{}

	if id.ID() != "admin" {
		t.Errorf("ID() = %q, want admin", id.ID())
	}
	if id.TenantID() != "admin" {
		t.Errorf("TenantID() = %q, want admin", id.TenantID())
	}
	if id.Email() != "" {
		t.Errorf("Email() = %q, want empty", id.Email())
	}
	if len(id.Scopes()) != 1 || id.Scopes()[0] != "admin" {
		t.Errorf("Scopes() = %v, want [admin]", id.Scopes())
	}
	if id.Roles() != nil {
		t.Error("Roles() should be nil")
	}
	if !id.HasScope("anything") {
		t.Error("HasScope should always return true")
	}
	if id.HasRole("admin") {
		t.Error("HasRole should always return false")
	}
	if id.Stats() != nil {
		t.Error("Stats() should be nil")
	}
}

// --- Credential endpoint integration tests ---

func TestSetAndGetCredential(t *testing.T) {
	handler := setupTestHandler(t)

	// Set a credential.
	body := `{"value":"ghp_secret123"}`
	req := httptest.NewRequest("PUT", "/credentials/agent-1/github-token", strings.NewReader(body))
	req.Header.Set("Authorization", authHeader())
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("PUT status = %d, want 200; body: %s", w.Code, w.Body.String())
	}

	var setResp CredentialResponse
	if err := json.NewDecoder(w.Body).Decode(&setResp); err != nil {
		t.Fatalf("decode PUT response: %v", err)
	}
	if setResp.Agent != "agent-1" || setResp.Key != "github-token" || setResp.Value != "ghp_secret123" {
		t.Errorf("unexpected PUT response: %+v", setResp)
	}

	// Get the credential back.
	req = httptest.NewRequest("GET", "/credentials/agent-1/github-token", nil)
	req.Header.Set("Authorization", authHeader())
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want 200; body: %s", w.Code, w.Body.String())
	}

	var getResp CredentialResponse
	if err := json.NewDecoder(w.Body).Decode(&getResp); err != nil {
		t.Fatalf("decode GET response: %v", err)
	}
	if getResp.Value != "ghp_secret123" {
		t.Errorf("GET value = %q, want ghp_secret123", getResp.Value)
	}
}

func TestListCredentials(t *testing.T) {
	handler := setupTestHandler(t)

	// Set two credentials for the same agent.
	for _, key := range []string{"github-token", "matrix-token"} {
		body := `{"value":"secret"}`
		req := httptest.NewRequest("PUT", "/credentials/agent-1/"+key, strings.NewReader(body))
		req.Header.Set("Authorization", authHeader())
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("PUT %s status = %d, want 200", key, w.Code)
		}
	}

	// List keys.
	req := httptest.NewRequest("GET", "/credentials/agent-1", nil)
	req.Header.Set("Authorization", authHeader())
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("LIST status = %d, want 200; body: %s", w.Code, w.Body.String())
	}

	var listResp CredentialKeyList
	if err := json.NewDecoder(w.Body).Decode(&listResp); err != nil {
		t.Fatalf("decode LIST response: %v", err)
	}
	if listResp.Agent != "agent-1" {
		t.Errorf("agent = %q, want agent-1", listResp.Agent)
	}
	if len(listResp.Keys) != 2 {
		t.Errorf("keys count = %d, want 2", len(listResp.Keys))
	}
}

func TestDeleteCredential(t *testing.T) {
	handler := setupTestHandler(t)

	// Set a credential.
	body := `{"value":"to-delete"}`
	req := httptest.NewRequest("PUT", "/credentials/agent-1/temp-key", strings.NewReader(body))
	req.Header.Set("Authorization", authHeader())
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("PUT status = %d, want 200", w.Code)
	}

	// Delete it.
	req = httptest.NewRequest("DELETE", "/credentials/agent-1/temp-key", nil)
	req.Header.Set("Authorization", authHeader())
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("DELETE status = %d, want 200; body: %s", w.Code, w.Body.String())
	}

	// Verify it's gone.
	req = httptest.NewRequest("GET", "/credentials/agent-1/temp-key", nil)
	req.Header.Set("Authorization", authHeader())
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("GET after DELETE status = %d, want 404", w.Code)
	}
}

func TestGetCredentialNotFound(t *testing.T) {
	handler := setupTestHandler(t)

	req := httptest.NewRequest("GET", "/credentials/agent-1/nonexistent", nil)
	req.Header.Set("Authorization", authHeader())
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}
