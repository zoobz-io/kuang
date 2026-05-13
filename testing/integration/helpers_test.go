//go:build testing

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zoobz-io/kuang"
	"github.com/zoobz-io/kuang/mcp"
)

// newMCPServer creates an MCP server connected to the test kuang server
// via mTLS using the given client certificate.
func newMCPServer(t *testing.T, env *kuang.TestEnv, certPath, keyPath string) *mcp.Server {
	t.Helper()
	client, err := mcp.NewClient(env.BaseURL, env.CACertPath, certPath, keyPath)
	if err != nil {
		t.Fatalf("mcp.NewClient: %v", err)
	}
	server := mcp.NewServer(client)
	if err := server.LoadTools(); err != nil {
		t.Fatalf("LoadTools: %v", err)
	}
	return server
}

// mcpRoundTrip sends a single JSON-RPC message and returns the parsed response.
func mcpRoundTrip(t *testing.T, s *mcp.Server, method string, id int, params string) map[string]any {
	t.Helper()
	msg := fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"method":"%s","params":%s}`, id, method, params)
	var out bytes.Buffer
	if err := s.Run(strings.NewReader(msg+"\n"), &out); err != nil {
		t.Fatalf("MCP Run: %v", err)
	}
	var resp map[string]any
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal MCP response: %v\nraw: %s", err, out.String())
	}
	return resp
}

// mcpToolCall invokes a tool and returns the text content and error flag.
func mcpToolCall(t *testing.T, s *mcp.Server, toolName string, args string, id int) (text string, isError bool) {
	t.Helper()
	resp := mcpRoundTrip(t, s, "tools/call", id,
		fmt.Sprintf(`{"name":"%s","arguments":%s}`, toolName, args))
	result, ok := resp["result"].(map[string]any)
	if !ok {
		if e, ok := resp["error"].(map[string]any); ok {
			return fmt.Sprintf("protocol error: %v", e["message"]), true
		}
		return "no result", true
	}
	isErr, _ := result["isError"].(bool)
	content, _ := result["content"].([]any)
	if len(content) > 0 {
		first, _ := content[0].(map[string]any)
		text, _ = first["text"].(string)
	}
	return text, isErr
}

// mcpToolNames returns the names of all tools from tools/list.
func mcpToolNames(t *testing.T, s *mcp.Server) []string {
	t.Helper()
	resp := mcpRoundTrip(t, s, "tools/list", 1, `{}`)
	result, _ := resp["result"].(map[string]any)
	tools, _ := result["tools"].([]any)
	var names []string
	for _, tool := range tools {
		tm, _ := tool.(map[string]any)
		name, _ := tm["name"].(string)
		names = append(names, name)
	}
	return names
}

// toolByName returns the tool name containing the given substring, or fails.
func toolByName(t *testing.T, names []string, substr string) string {
	t.Helper()
	for _, n := range names {
		if strings.Contains(strings.ToLower(n), strings.ToLower(substr)) {
			return n
		}
	}
	t.Fatalf("no tool matching %q in %v", substr, names)
	return ""
}

// hasToolMatching returns true if any tool name contains the substring.
func hasToolMatching(names []string, substr string) bool {
	for _, n := range names {
		if strings.Contains(strings.ToLower(n), strings.ToLower(substr)) {
			return true
		}
	}
	return false
}

// fullAccessAgent creates an MCP server with all github scopes.
func fullAccessAgent(t *testing.T, env *kuang.TestEnv, allScopes []string) *mcp.Server {
	t.Helper()
	cert, key := env.IssueClient(t, "full-agent", allScopes)
	return newMCPServer(t, env, cert, key)
}

// issueAgent creates an MCP server from a client cert at known paths.
func issueAgent(t *testing.T, env *kuang.TestEnv, name string, scopes []string) *mcp.Server {
	t.Helper()
	cert, key := env.IssueClient(t, name, scopes)
	return newMCPServer(t, env, cert, key)
}

// certDir returns the directory portion of the CA cert path (for convenience).
func certDir(env *kuang.TestEnv) string {
	return filepath.Dir(env.CACertPath)
}
