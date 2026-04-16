package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// mockBackend implements Backend for testing.
type mockBackend struct {
	spec    map[string]any
	callFn  func(ctx context.Context, op Operation, args json.RawMessage) (int, []byte, error)
}

func (m *mockBackend) FetchSpec(_ context.Context) (map[string]any, error) {
	if m.spec == nil {
		return nil, fmt.Errorf("no spec")
	}
	return m.spec, nil
}

func (m *mockBackend) Call(ctx context.Context, op Operation, args json.RawMessage) (int, []byte, error) {
	if m.callFn != nil {
		return m.callFn(ctx, op, args)
	}
	return 200, []byte(`{"ok":true}`), nil
}

func newTestServer(tools []Tool, opMap map[string]Operation) *Server {
	return &Server{
		backend: &mockBackend{},
		tools:   tools,
		opMap:   opMap,
	}
}

func marshalResult(t *testing.T, result any) []byte {
	t.Helper()
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal result: %v", err)
	}
	return data
}

func roundTrip(t *testing.T, s *Server, request string) Response {
	t.Helper()
	var out bytes.Buffer
	in := strings.NewReader(request + "\n")
	if err := s.Run(in, &out); err != nil {
		t.Fatalf("Run error: %v", err)
	}
	var resp Response
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v\nraw: %s", err, out.String())
	}
	return resp
}

func TestInitialize(t *testing.T) {
	s := newTestServer(nil, nil)
	resp := roundTrip(t, s, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25"}}`)

	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error.Message)
	}

	data := marshalResult(t, resp.Result)
	var result InitializeResult
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}

	if result.ProtocolVersion != protocolVersion {
		t.Errorf("version = %q, want %q", result.ProtocolVersion, protocolVersion)
	}
	if result.ServerInfo.Name != "kuang-mcp" {
		t.Errorf("name = %q, want kuang-mcp", result.ServerInfo.Name)
	}
	if result.Capabilities.Tools == nil {
		t.Error("expected tools capability")
	}
}

func TestToolsList(t *testing.T) {
	tools := []Tool{
		{Name: "get_users", Description: "List users", InputSchema: json.RawMessage(`{"type":"object"}`)},
		{Name: "create_user", Description: "Create user", InputSchema: json.RawMessage(`{"type":"object"}`)},
	}
	s := newTestServer(tools, nil)
	resp := roundTrip(t, s, `{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`)

	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error.Message)
	}

	data := marshalResult(t, resp.Result)
	var result ToolsListResult
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}

	if len(result.Tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(result.Tools))
	}
	if result.Tools[0].Name != "get_users" {
		t.Errorf("first tool = %q, want get_users", result.Tools[0].Name)
	}
}

func TestToolsCallUnknown(t *testing.T) {
	s := newTestServer(nil, map[string]Operation{})
	resp := roundTrip(t, s, `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"nope","arguments":{}}}`)

	if resp.Error == nil {
		t.Fatal("expected error for unknown tool")
	}
	if resp.Error.Code != -32602 {
		t.Errorf("code = %d, want -32602", resp.Error.Code)
	}
}

func TestToolsCallSuccess(t *testing.T) {
	mock := &mockBackend{
		callFn: func(_ context.Context, _ Operation, _ json.RawMessage) (int, []byte, error) {
			return 200, []byte(`{"users":["alice","bob"]}`), nil
		},
	}
	s := &Server{
		backend: mock,
		tools:   []Tool{{Name: "list_users"}},
		opMap:   map[string]Operation{"list_users": {Method: "GET", Path: "/users"}},
	}
	resp := roundTrip(t, s, `{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"list_users","arguments":{}}}`)

	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error.Message)
	}

	data := marshalResult(t, resp.Result)
	var result ToolCallResult
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if result.IsError {
		t.Error("expected isError=false")
	}
	if len(result.Content) != 1 || result.Content[0].Text != `{"users":["alice","bob"]}` {
		t.Errorf("content = %+v", result.Content)
	}
}

func TestToolsCallHTTPError(t *testing.T) {
	mock := &mockBackend{
		callFn: func(_ context.Context, _ Operation, _ json.RawMessage) (int, []byte, error) {
			return 403, []byte(`{"code":"forbidden","message":"nope"}`), nil
		},
	}
	s := &Server{
		backend: mock,
		tools:   []Tool{{Name: "delete_user"}},
		opMap:   map[string]Operation{"delete_user": {Method: "DELETE", Path: "/users/{id}"}},
	}
	resp := roundTrip(t, s, `{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"delete_user","arguments":{"id":"123"}}}`)

	if resp.Error != nil {
		t.Fatalf("unexpected protocol error: %s", resp.Error.Message)
	}

	data := marshalResult(t, resp.Result)
	var result ToolCallResult
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !result.IsError {
		t.Error("expected isError=true for 403")
	}
}

func TestToolsCallNetworkError(t *testing.T) {
	mock := &mockBackend{
		callFn: func(_ context.Context, _ Operation, _ json.RawMessage) (int, []byte, error) {
			return 0, nil, fmt.Errorf("connection refused")
		},
	}
	s := &Server{
		backend: mock,
		tools:   []Tool{{Name: "get_health"}},
		opMap:   map[string]Operation{"get_health": {Method: "GET", Path: "/health"}},
	}
	resp := roundTrip(t, s, `{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"get_health","arguments":{}}}`)

	if resp.Error != nil {
		t.Fatalf("unexpected protocol error: %s", resp.Error.Message)
	}

	data := marshalResult(t, resp.Result)
	var result ToolCallResult
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !result.IsError {
		t.Error("expected isError=true for network error")
	}
	if !strings.Contains(result.Content[0].Text, "connection refused") {
		t.Errorf("expected error message, got %q", result.Content[0].Text)
	}
}

func TestToolsCallEmptyBody(t *testing.T) {
	mock := &mockBackend{
		callFn: func(_ context.Context, _ Operation, _ json.RawMessage) (int, []byte, error) {
			return 204, nil, nil
		},
	}
	s := &Server{
		backend: mock,
		tools:   []Tool{{Name: "do_thing"}},
		opMap:   map[string]Operation{"do_thing": {Method: "POST", Path: "/thing"}},
	}
	resp := roundTrip(t, s, `{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"do_thing","arguments":{}}}`)

	data := marshalResult(t, resp.Result)
	var result ToolCallResult
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if result.Content[0].Text != "HTTP 204" {
		t.Errorf("expected 'HTTP 204', got %q", result.Content[0].Text)
	}
}

func TestToolsCallInvalidParams(t *testing.T) {
	s := newTestServer(nil, map[string]Operation{})
	resp := roundTrip(t, s, `{"jsonrpc":"2.0","id":8,"method":"tools/call","params":"bad"}`)

	if resp.Error == nil {
		t.Fatal("expected error for invalid params")
	}
	if resp.Error.Code != -32602 {
		t.Errorf("code = %d, want -32602", resp.Error.Code)
	}
}

func TestLoadTools(t *testing.T) {
	mock := &mockBackend{
		spec: map[string]any{
			"paths": map[string]any{
				"/users": map[string]any{
					"get": map[string]any{
						"operationId": "listUsers",
						"summary":     "List users",
					},
				},
			},
		},
	}
	s := NewServer(mock)
	if err := s.LoadTools(); err != nil {
		t.Fatalf("LoadTools: %v", err)
	}
	if s.ToolCount() != 1 {
		t.Errorf("ToolCount = %d, want 1", s.ToolCount())
	}
}

func TestLoadToolsFetchError(t *testing.T) {
	mock := &mockBackend{spec: nil}
	s := NewServer(mock)
	if err := s.LoadTools(); err == nil {
		t.Fatal("expected error when spec fetch fails")
	}
}

func TestPing(t *testing.T) {
	s := newTestServer(nil, nil)
	resp := roundTrip(t, s, `{"jsonrpc":"2.0","id":9,"method":"ping","params":{}}`)

	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error.Message)
	}
}

func TestMethodNotFound(t *testing.T) {
	s := newTestServer(nil, nil)
	resp := roundTrip(t, s, `{"jsonrpc":"2.0","id":10,"method":"unknown/method","params":{}}`)

	if resp.Error == nil {
		t.Fatal("expected error for unknown method")
	}
	if resp.Error.Code != -32601 {
		t.Errorf("code = %d, want -32601", resp.Error.Code)
	}
}

func TestInitializedNotification(t *testing.T) {
	s := newTestServer(nil, nil)
	var out bytes.Buffer
	in := strings.NewReader(`{"jsonrpc":"2.0","method":"notifications/initialized"}` + "\n")
	if err := s.Run(in, &out); err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("expected no response for notification, got: %s", out.String())
	}
}
