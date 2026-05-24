//go:build testing

package admin

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/zoobz-io/kuang/internal/auth"
	"github.com/zoobz-io/kuang/internal/creds"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

const (
	testUsername = "admin"
	testPassword = "test-password-12345"
)

func setupTestHandler(t *testing.T) http.Handler {
	t.Helper()
	sum.Reset()
	dbPath := filepath.Join(t.TempDir(), "test.db")

	store, err := creds.Open(dbPath)
	if err != nil {
		t.Fatalf("open test store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	userStore, err := auth.OpenStore(dbPath)
	if err != nil {
		t.Fatalf("open user store: %v", err)
	}
	t.Cleanup(func() { _ = userStore.Close() })

	if err := userStore.Create(t.Context(), testUsername, testPassword); err != nil {
		t.Fatalf("create test user: %v", err)
	}

	return NewHandler(store, LocalAuthenticator(userStore), userStore)
}

func basicAuth(username, password string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))
}

func authHeader() string {
	return basicAuth(testUsername, testPassword)
}

// --- Auth tests ---

func TestAuthenticateRejectsNoHeader(t *testing.T) {
	handler := setupTestHandler(t)
	req := httptest.NewRequest("GET", "/credentials/agent-1", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", w.Code)
	}
}

func TestAuthenticateRejectsWrongPassword(t *testing.T) {
	handler := setupTestHandler(t)
	req := httptest.NewRequest("GET", "/credentials/agent-1", nil)
	req.Header.Set("Authorization", basicAuth(testUsername, "wrong-password"))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", w.Code)
	}
}

func TestAuthenticateAcceptsValidCredentials(t *testing.T) {
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
	id := &AdminIdentity{username: "testuser", scopes: []string{"admin"}}

	if id.ID() != "testuser" {
		t.Errorf("ID() = %q, want testuser", id.ID())
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
	if !id.HasScope("admin") {
		t.Error("HasScope(admin) should return true")
	}
	if id.HasScope("other") {
		t.Error("HasScope(other) should return false")
	}
	if id.HasRole("admin") {
		t.Error("HasRole should always return false")
	}
	if id.Stats() != nil {
		t.Error("Stats() should be nil")
	}
}

// --- Custom authenticator test ---

func TestCustomAuthenticator(t *testing.T) {
	sum.Reset()
	dbPath := filepath.Join(t.TempDir(), "test.db")

	store, err := creds.Open(dbPath)
	if err != nil {
		t.Fatalf("open test store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	// Custom authenticator that accepts a Bearer token.
	custom := func(_ context.Context, r *http.Request) (rocco.Identity, error) {
		if r.Header.Get("Authorization") == "Bearer magic-token" {
			return &AdminIdentity{username: "custom-user", scopes: []string{ScopeAdmin}}, nil
		}
		return rocco.NoIdentity{}, nil
	}

	handler := NewHandler(store, custom, nil)

	// Valid token works.
	req := httptest.NewRequest("GET", "/credentials/agent-1", nil)
	req.Header.Set("Authorization", "Bearer magic-token")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}

	// Invalid token rejected.
	req = httptest.NewRequest("GET", "/credentials/agent-1", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", w.Code)
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

// --- Setup endpoint tests ---

func setupEmptyHandler(t *testing.T) (http.Handler, *auth.Store) {
	t.Helper()
	sum.Reset()
	dbPath := filepath.Join(t.TempDir(), "test.db")

	store, err := creds.Open(dbPath)
	if err != nil {
		t.Fatalf("open test store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	userStore, err := auth.OpenStore(dbPath)
	if err != nil {
		t.Fatalf("open user store: %v", err)
	}
	t.Cleanup(func() { _ = userStore.Close() })

	handler := NewHandler(store, LocalAuthenticator(userStore), userStore)
	return handler, userStore
}

func TestSetupCreatesFirstUser(t *testing.T) {
	handler, _ := setupEmptyHandler(t)

	body := `{"username":"myadmin","password":"secure-password-123"}`
	req := httptest.NewRequest("POST", "/setup", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("setup status = %d, want 200; body: %s", w.Code, w.Body.String())
	}

	var resp SetupResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode setup response: %v", err)
	}
	if resp.Username != "myadmin" {
		t.Errorf("username = %q, want myadmin", resp.Username)
	}

	// Auth works with the created user.
	req = httptest.NewRequest("GET", "/credentials/agent-1", nil)
	req.Header.Set("Authorization", basicAuth("myadmin", "secure-password-123"))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("auth after setup status = %d, want 200; body: %s", w.Code, w.Body.String())
	}
}

func TestSetupRejectsSecondCall(t *testing.T) {
	handler, _ := setupEmptyHandler(t)

	// First call succeeds.
	body := `{"username":"myadmin","password":"secure-password-123"}`
	req := httptest.NewRequest("POST", "/setup", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("first setup status = %d, want 200", w.Code)
	}

	// Second call is rejected.
	req = httptest.NewRequest("POST", "/setup", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("second setup status = %d, want 404; body: %s", w.Code, w.Body.String())
	}
}

func TestSetupRejectsShortPassword(t *testing.T) {
	handler, _ := setupEmptyHandler(t)

	body := `{"username":"myadmin","password":"short"}`
	req := httptest.NewRequest("POST", "/setup", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Error("setup should reject short passwords")
	}
}

func TestSetupNotAvailableWithCustomAuth(t *testing.T) {
	sum.Reset()
	dbPath := filepath.Join(t.TempDir(), "test.db")

	store, err := creds.Open(dbPath)
	if err != nil {
		t.Fatalf("open test store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	custom := func(_ context.Context, r *http.Request) (rocco.Identity, error) {
		return &AdminIdentity{username: "ext", scopes: []string{ScopeAdmin}}, nil
	}

	// nil userStore = no setup endpoint.
	handler := NewHandler(store, custom, nil)

	body := `{"username":"myadmin","password":"secure-password-123"}`
	req := httptest.NewRequest("POST", "/setup", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Error("setup should not be available with custom authenticator")
	}
}

func TestSetupUnavailableWhenAlreadyInitialized(t *testing.T) {
	handler, userStore := setupEmptyHandler(t)

	// Initialize directly via the store.
	if err := userStore.Initialize(t.Context(), "existing", "existing-password-123"); err != nil {
		t.Fatalf("initialize store: %v", err)
	}

	body := `{"username":"myadmin","password":"secure-password-123"}`
	req := httptest.NewRequest("POST", "/setup", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("setup when initialized status = %d, want 404; body: %s", w.Code, w.Body.String())
	}
}
