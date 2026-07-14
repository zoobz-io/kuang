package mcp

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestClientFetchSpec(t *testing.T) {
	spec := map[string]any{
		"openapi": "3.0.0",
		"paths": map[string]any{
			"/users": map[string]any{},
		},
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/openapi" {
			t.Errorf("path = %q, want /openapi", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(spec)
	}))
	defer ts.Close()

	c := &Client{http: ts.Client(), baseURL: ts.URL}
	got, err := c.FetchSpec(context.Background())
	if err != nil {
		t.Fatalf("FetchSpec: %v", err)
	}
	if got["openapi"] != "3.0.0" {
		t.Errorf("openapi = %v, want 3.0.0", got["openapi"])
	}
	if _, ok := got["paths"].(map[string]any); !ok {
		t.Errorf("missing paths in spec: %v", got)
	}
}

func TestClientFetchSpecError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, "boom")
	}))
	defer ts.Close()

	c := &Client{http: ts.Client(), baseURL: ts.URL}
	if _, err := c.FetchSpec(context.Background()); err == nil {
		t.Fatal("expected error for non-200 status")
	}
}

func TestClientCall(t *testing.T) {
	var (
		gotMethod string
		gotPath   string
		gotQuery  string
		gotBody   []byte
		gotCT     string
	)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		gotCT = r.Header.Get("Content-Type")
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"created":true}`)
	}))
	defer ts.Close()

	c := &Client{http: ts.Client(), baseURL: ts.URL}

	op := Operation{
		Method: http.MethodPost,
		Path:   "/users/{id}/posts",
		Parameters: []Parameter{
			{Name: "id", In: "path"},
			{Name: "verbose", In: "query"},
		},
	}
	args := json.RawMessage(`{"id":"42","verbose":"true","body":{"title":"hello"}}`)

	status, respBody, err := c.Call(context.Background(), op, args)
	if err != nil {
		t.Fatalf("Call: %v", err)
	}

	if status != http.StatusCreated {
		t.Errorf("status = %d, want 201", status)
	}
	if string(respBody) != `{"created":true}` {
		t.Errorf("body = %q, want {\"created\":true}", respBody)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotPath != "/users/42/posts" {
		t.Errorf("path = %q, want /users/42/posts", gotPath)
	}
	if gotQuery != "verbose=true" {
		t.Errorf("query = %q, want verbose=true", gotQuery)
	}
	if gotCT != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", gotCT)
	}

	var sentBody map[string]any
	if err := json.Unmarshal(gotBody, &sentBody); err != nil {
		t.Fatalf("unmarshal sent body: %v (raw %q)", err, gotBody)
	}
	want := map[string]any{"title": "hello"}
	if !reflect.DeepEqual(sentBody, want) {
		t.Errorf("sent body = %v, want %v", sentBody, want)
	}
}

func TestClientCallNoBody(t *testing.T) {
	var gotBody []byte
	var gotCT string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		gotCT = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := &Client{http: ts.Client(), baseURL: ts.URL}
	op := Operation{Method: http.MethodGet, Path: "/health"}

	status, _, err := c.Call(context.Background(), op, nil)
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if status != http.StatusOK {
		t.Errorf("status = %d, want 200", status)
	}
	if len(gotBody) != 0 {
		t.Errorf("expected empty body, got %q", gotBody)
	}
	if gotCT != "" {
		t.Errorf("expected no Content-Type, got %q", gotCT)
	}
}

func TestNewClientBadCertPaths(t *testing.T) {
	_, err := NewClient(Config{
		Host: "https://example.com",
		CA:   "/nonexistent/ca.pem",
		Cert: "/nonexistent/cert.pem",
		Key:  "/nonexistent/key.pem",
	})
	if err == nil {
		t.Fatal("expected error for missing cert paths")
	}
}
