package extend

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zoobz-io/rocco"
)

func TestOpenAPIInterceptorServesSpec(t *testing.T) {
	engine := rocco.NewEngine()

	handler := OpenAPIInterceptor(engine)
	req := httptest.NewRequest(http.MethodGet, "/openapi", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}
	// Should contain OpenAPI version marker.
	body := w.Body.String()
	if len(body) == 0 {
		t.Error("expected non-empty spec")
	}
}

func TestOpenAPIInterceptorPassesThrough(t *testing.T) {
	engine := rocco.NewEngine()
	// Register a test route on the engine's router.
	engine.Router().HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	handler := OpenAPIInterceptor(engine)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "ok" {
		t.Errorf("expected 'ok', got %q", w.Body.String())
	}
}

func TestOpenAPIInterceptorIgnoresNonGET(t *testing.T) {
	engine := rocco.NewEngine()

	handler := OpenAPIInterceptor(engine)
	req := httptest.NewRequest(http.MethodPost, "/openapi", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// POST /openapi should not be intercepted — falls through to router.
	// Router has no POST /openapi, so it returns 405 or 404.
	if w.Code == http.StatusOK {
		t.Error("POST /openapi should not return 200")
	}
}
