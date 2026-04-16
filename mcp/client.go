package mcp

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// Client makes mTLS HTTP calls to a kuang server.
type Client struct {
	http    *http.Client
	baseURL string
}

// NewClient creates an HTTP client configured for mTLS.
func NewClient(baseURL, caCertPath, certPath, keyPath string) (*Client, error) {
	caPEM, err := os.ReadFile(caCertPath) //nolint:gosec // cert paths are user-configured
	if err != nil {
		return nil, fmt.Errorf("read CA cert: %w", err)
	}
	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caPEM) {
		return nil, fmt.Errorf("failed to parse CA cert")
	}

	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, fmt.Errorf("load client keypair: %w", err)
	}

	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					Certificates: []tls.Certificate{cert},
					RootCAs:      caPool,
					MinVersion:   tls.VersionTLS13,
				},
			},
		},
	}, nil
}

// FetchSpec retrieves the OpenAPI spec from the kuang server.
func (c *Client) FetchSpec(ctx context.Context) (map[string]any, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/openapi", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch openapi: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openapi returned %d: %s", resp.StatusCode, body)
	}

	var spec map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&spec); err != nil {
		return nil, fmt.Errorf("decode openapi: %w", err)
	}
	return spec, nil
}

// Call makes an HTTP request to the kuang server for a given operation.
func (c *Client) Call(ctx context.Context, op Operation, arguments json.RawMessage) (int, []byte, error) {
	var args map[string]any
	if arguments != nil {
		if err := json.Unmarshal(arguments, &args); err != nil {
			return 0, nil, fmt.Errorf("parse arguments: %w", err)
		}
	}

	// Build URL with path and query parameters.
	path := op.Path
	query := ""
	for _, p := range op.Parameters {
		val, ok := args[p.Name]
		if !ok {
			continue
		}
		s := fmt.Sprintf("%v", val)
		switch p.In {
		case "path":
			path = strings.ReplaceAll(path, "{"+p.Name+"}", s)
		case "query":
			if query != "" {
				query += "&"
			}
			query += p.Name + "=" + s
		}
	}

	url := c.baseURL + path
	if query != "" {
		url += "?" + query
	}

	// Build request body from the "body" argument.
	var bodyReader io.Reader
	if bodyVal, ok := args["body"]; ok {
		data, err := json.Marshal(bodyVal)
		if err != nil {
			return 0, nil, fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, op.Method, url, bodyReader)
	if err != nil {
		return 0, nil, fmt.Errorf("create request: %w", err)
	}
	if bodyReader != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("read response: %w", err)
	}

	return resp.StatusCode, respBody, nil
}
