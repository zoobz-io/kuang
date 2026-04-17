// Package httpc provides an instrumented HTTP client for external API calls.
// Every request emits capitan signals for observability and can be composed
// with pipz processors for retry, circuit breaking, and rate limiting.
package httpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/zoobz-io/capitan"
)

// Client is an instrumented HTTP client for external service calls.
type Client struct {
	headers map[string]string
	http    *http.Client
	baseURL string
}

// Option configures a Client.
type Option func(*Client)

// WithBaseURL sets the base URL for all requests.
func WithBaseURL(url string) Option {
	return func(c *Client) { c.baseURL = url }
}

// WithHeader adds a default header to all requests.
func WithHeader(key, value string) Option {
	return func(c *Client) { c.headers[key] = value }
}

// WithBearerToken adds a Bearer authorization header.
func WithBearerToken(token string) Option {
	return WithHeader("Authorization", "Bearer "+token)
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.http.Timeout = d }
}

// New creates an instrumented HTTP client.
func New(opts ...Option) *Client {
	c := &Client{
		http:    &http.Client{Timeout: 30 * time.Second},
		headers: make(map[string]string),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Response wraps an HTTP response with the decoded body.
type Response struct {
	Headers    http.Header
	Body       []byte
	DurationMs int64
	Status     int
}

// Decode unmarshals the response body into the given value.
func (r *Response) Decode(v any) error {
	return json.Unmarshal(r.Body, v)
}

// Get performs a GET request.
func (c *Client) Get(ctx context.Context, path string) (*Response, error) {
	return c.Do(ctx, http.MethodGet, path, nil)
}

// Post performs a POST request with a JSON body.
func (c *Client) Post(ctx context.Context, path string, body any) (*Response, error) {
	return c.Do(ctx, http.MethodPost, path, body)
}

// Put performs a PUT request with a JSON body.
func (c *Client) Put(ctx context.Context, path string, body any) (*Response, error) {
	return c.Do(ctx, http.MethodPut, path, body)
}

// Patch performs a PATCH request with a JSON body.
func (c *Client) Patch(ctx context.Context, path string, body any) (*Response, error) {
	return c.Do(ctx, http.MethodPatch, path, body)
}

// Delete performs a DELETE request.
func (c *Client) Delete(ctx context.Context, path string) (*Response, error) {
	return c.Do(ctx, http.MethodDelete, path, nil)
}

// Do performs an HTTP request with instrumentation.
func (c *Client) Do(ctx context.Context, method, path string, body any) (*Response, error) {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	capitan.Info(ctx, RequestStarted,
		MethodKey.Field(method),
		URLKey.Field(url),
	)

	start := time.Now()
	resp, err := c.http.Do(req)
	duration := time.Since(start)

	if err != nil {
		capitan.Error(ctx, RequestFailed,
			MethodKey.Field(method),
			URLKey.Field(url),
			DurationMsKey.Field(duration.Milliseconds()),
			ErrorKey.Field(err.Error()),
		)
		return nil, fmt.Errorf("%s %s: %w", method, path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	r := &Response{
		Status:     resp.StatusCode,
		Body:       respBody,
		Headers:    resp.Header,
		DurationMs: duration.Milliseconds(),
	}

	if resp.StatusCode >= 400 {
		capitan.Warn(ctx, RequestError,
			MethodKey.Field(method),
			URLKey.Field(url),
			StatusKey.Field(resp.StatusCode),
			DurationMsKey.Field(duration.Milliseconds()),
		)
		return r, fmt.Errorf("%s %s returned %d: %s", method, path, resp.StatusCode, respBody)
	}

	capitan.Info(ctx, RequestCompleted,
		MethodKey.Field(method),
		URLKey.Field(url),
		StatusKey.Field(resp.StatusCode),
		DurationMsKey.Field(duration.Milliseconds()),
	)

	return r, nil
}
