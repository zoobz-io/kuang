package httpc

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// recordingServer captures details of the last received request.
type recordingServer struct {
	method      string
	path        string
	body        []byte
	contentType string
	accept      string
	authz       string
	header      http.Header
}

func newRecordingServer(t *testing.T, status int, respBody string) (*httptest.Server, *recordingServer) {
	t.Helper()
	rec := &recordingServer{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec.method = r.Method
		rec.path = r.URL.Path
		rec.body, _ = io.ReadAll(r.Body)
		rec.contentType = r.Header.Get("Content-Type")
		rec.accept = r.Header.Get("Accept")
		rec.authz = r.Header.Get("Authorization")
		rec.header = r.Header.Clone()
		w.WriteHeader(status)
		_, _ = io.WriteString(w, respBody)
	}))
	t.Cleanup(ts.Close)
	return ts, rec
}

func TestClientMethods(t *testing.T) {
	tests := []struct {
		name     string
		hasBody  bool
		wantVerb string
		call     func(c *Client, ctx context.Context) (*Response, error)
	}{
		{
			name:     "Get",
			wantVerb: http.MethodGet,
			call: func(c *Client, ctx context.Context) (*Response, error) {
				return c.Get(ctx, "/thing")
			},
		},
		{
			name:     "Post",
			hasBody:  true,
			wantVerb: http.MethodPost,
			call: func(c *Client, ctx context.Context) (*Response, error) {
				return c.Post(ctx, "/thing", map[string]string{"k": "v"})
			},
		},
		{
			name:     "Put",
			hasBody:  true,
			wantVerb: http.MethodPut,
			call: func(c *Client, ctx context.Context) (*Response, error) {
				return c.Put(ctx, "/thing", map[string]string{"k": "v"})
			},
		},
		{
			name:     "Patch",
			hasBody:  true,
			wantVerb: http.MethodPatch,
			call: func(c *Client, ctx context.Context) (*Response, error) {
				return c.Patch(ctx, "/thing", map[string]string{"k": "v"})
			},
		},
		{
			name:     "Delete",
			wantVerb: http.MethodDelete,
			call: func(c *Client, ctx context.Context) (*Response, error) {
				return c.Delete(ctx, "/thing")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, rec := newRecordingServer(t, http.StatusOK, `{"ok":true}`)
			c := New(WithBaseURL(ts.URL))

			resp, err := tt.call(c, context.Background())
			if err != nil {
				t.Fatalf("%s: %v", tt.name, err)
			}
			if resp.Status != http.StatusOK {
				t.Errorf("status = %d, want 200", resp.Status)
			}
			if rec.method != tt.wantVerb {
				t.Errorf("method = %q, want %q", rec.method, tt.wantVerb)
			}
			if rec.path != "/thing" {
				t.Errorf("path = %q, want /thing", rec.path)
			}
			if rec.accept != "application/json" {
				t.Errorf("Accept = %q, want application/json", rec.accept)
			}
			if tt.hasBody {
				if rec.contentType != "application/json" {
					t.Errorf("Content-Type = %q, want application/json", rec.contentType)
				}
				if !strings.Contains(string(rec.body), `"k":"v"`) {
					t.Errorf("body = %q, want to contain \"k\":\"v\"", rec.body)
				}
			} else {
				if len(rec.body) != 0 {
					t.Errorf("expected empty body, got %q", rec.body)
				}
				if rec.contentType != "" {
					t.Errorf("Content-Type = %q, want empty for bodyless request", rec.contentType)
				}
			}
		})
	}
}

func TestResponseDecode(t *testing.T) {
	ts, _ := newRecordingServer(t, http.StatusOK, `{"name":"alice","age":30}`)
	c := New(WithBaseURL(ts.URL))

	resp, err := c.Get(context.Background(), "/user")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	var out struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	if err := resp.Decode(&out); err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if out.Name != "alice" {
		t.Errorf("name = %q, want alice", out.Name)
	}
	if out.Age != 30 {
		t.Errorf("age = %d, want 30", out.Age)
	}
}

func TestWithRequestBearerToken(t *testing.T) {
	ts, rec := newRecordingServer(t, http.StatusOK, `{}`)
	c := New(WithBaseURL(ts.URL))

	if _, err := c.Get(context.Background(), "/x", WithRequestBearerToken("tok123")); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if rec.authz != "Bearer tok123" {
		t.Errorf("Authorization = %q, want Bearer tok123", rec.authz)
	}
}

func TestRequestOptionOverridesClientDefaults(t *testing.T) {
	ts, rec := newRecordingServer(t, http.StatusOK, `{}`)
	c := New(
		WithBaseURL(ts.URL),
		WithBearerToken("client-default"),
		WithHeader("X-Custom", "client-value"),
	)

	_, err := c.Get(context.Background(), "/x",
		WithRequestBearerToken("per-request"),
		WithRequestHeader("X-Custom", "req-value"),
	)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if rec.authz != "Bearer per-request" {
		t.Errorf("Authorization = %q, want per-request override", rec.authz)
	}
	if got := rec.header.Get("X-Custom"); got != "req-value" {
		t.Errorf("X-Custom = %q, want req-value", got)
	}
}

func TestClientLevelDefaultsApplied(t *testing.T) {
	ts, rec := newRecordingServer(t, http.StatusOK, `{}`)
	c := New(
		WithBaseURL(ts.URL),
		WithBearerToken("client-default"),
		WithHeader("X-Custom", "client-value"),
	)

	if _, err := c.Get(context.Background(), "/x"); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if rec.authz != "Bearer client-default" {
		t.Errorf("Authorization = %q, want Bearer client-default", rec.authz)
	}
	if got := rec.header.Get("X-Custom"); got != "client-value" {
		t.Errorf("X-Custom = %q, want client-value", got)
	}
}

func TestAcceptDefaultsToJSON(t *testing.T) {
	ts, rec := newRecordingServer(t, http.StatusOK, `{}`)
	c := New(WithBaseURL(ts.URL))

	if _, err := c.Get(context.Background(), "/x"); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if rec.accept != "application/json" {
		t.Errorf("Accept = %q, want application/json (default)", rec.accept)
	}
}

func TestAcceptHeaderNotClobbered(t *testing.T) {
	// A client-level or per-request Accept must survive; the default must not
	// overwrite a caller-provided media type (e.g. a module's own).
	t.Run("client-level", func(t *testing.T) {
		ts, rec := newRecordingServer(t, http.StatusOK, `{}`)
		c := New(WithBaseURL(ts.URL), WithHeader("Accept", "application/vnd.github+json"))
		if _, err := c.Get(context.Background(), "/x"); err != nil {
			t.Fatalf("Get: %v", err)
		}
		if rec.accept != "application/vnd.github+json" {
			t.Errorf("Accept = %q, want application/vnd.github+json", rec.accept)
		}
	})
	t.Run("per-request", func(t *testing.T) {
		ts, rec := newRecordingServer(t, http.StatusOK, `{}`)
		c := New(WithBaseURL(ts.URL))
		_, err := c.Get(context.Background(), "/x", WithRequestHeader("Accept", "text/plain"))
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if rec.accept != "text/plain" {
			t.Errorf("Accept = %q, want text/plain", rec.accept)
		}
	})
}

func TestErrorStatusReturnsResponseAndError(t *testing.T) {
	ts, _ := newRecordingServer(t, http.StatusNotFound, `{"error":"not found here"}`)
	c := New(WithBaseURL(ts.URL))

	resp, err := c.Get(context.Background(), "/missing")
	if err == nil {
		t.Fatal("expected error for status 404")
	}
	if resp == nil {
		t.Fatal("expected non-nil Response alongside error")
	}
	if resp.Status != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.Status)
	}
	if !strings.Contains(err.Error(), "not found here") {
		t.Errorf("error %q should include response body", err.Error())
	}
	if string(resp.Body) != `{"error":"not found here"}` {
		t.Errorf("body = %q", resp.Body)
	}
}

func TestTransportError(t *testing.T) {
	ts, _ := newRecordingServer(t, http.StatusOK, `{}`)
	url := ts.URL
	ts.Close() // close so connections are refused

	c := New(WithBaseURL(url))
	resp, err := c.Get(context.Background(), "/x")
	if err == nil {
		t.Fatal("expected transport error against closed server")
	}
	if resp != nil {
		t.Errorf("expected nil Response on transport error, got %+v", resp)
	}
}

func TestMarshalBodyError(t *testing.T) {
	c := New(WithBaseURL("http://example.invalid"))
	// A channel cannot be JSON-marshaled.
	_, err := c.Post(context.Background(), "/x", make(chan int))
	if err == nil {
		t.Fatal("expected marshal error for unmarshalable body")
	}
}
