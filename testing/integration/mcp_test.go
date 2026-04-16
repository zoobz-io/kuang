//go:build testing

package integration

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/zoobz-io/kuang/config"
	"github.com/zoobz-io/kuang/internal/auth"
	"github.com/zoobz-io/kuang/mcp"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sctx"
)

// --- PKI helpers ---

type testCA struct {
	dir    string
	key    ed25519.PrivateKey
	cert   *x509.Certificate
	serial int64
}

func newTestCA(t *testing.T) *testCA {
	t.Helper()
	dir := t.TempDir()

	pub, key, _ := ed25519.GenerateKey(rand.Reader)
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{Organization: []string{"Kuang Test CA"}},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}
	der, _ := x509.CreateCertificate(rand.Reader, template, template, pub, key)
	cert, _ := x509.ParseCertificate(der)
	writePEM(t, filepath.Join(dir, "ca.pem"), "CERTIFICATE", der)

	return &testCA{dir: dir, key: key, cert: cert, serial: 1}
}

func (ca *testCA) issueServer(t *testing.T, cn string, permissions []string) {
	t.Helper()
	ca.serial++
	pub, key, _ := ed25519.GenerateKey(rand.Reader)
	template := &x509.Certificate{
		SerialNumber: big.NewInt(ca.serial),
		Subject: pkix.Name{
			CommonName:         cn,
			OrganizationalUnit: permissions,
		},
		NotBefore:   time.Now().Add(-time.Hour),
		NotAfter:    time.Now().Add(24 * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageAny},
		DNSNames:    []string{"localhost"},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1)},
	}
	der, _ := x509.CreateCertificate(rand.Reader, template, ca.cert, pub, ca.key)
	writePEM(t, filepath.Join(ca.dir, "server.pem"), "CERTIFICATE", der)
	keyBytes, _ := x509.MarshalPKCS8PrivateKey(key)
	writePEM(t, filepath.Join(ca.dir, "server-key.pem"), "PRIVATE KEY", keyBytes)
}

func (ca *testCA) issueClient(t *testing.T, name string, permissions []string) (certPath, keyPath string) {
	t.Helper()
	ca.serial++
	pub, key, _ := ed25519.GenerateKey(rand.Reader)
	template := &x509.Certificate{
		SerialNumber: big.NewInt(ca.serial),
		Subject: pkix.Name{
			CommonName:         name,
			OrganizationalUnit: permissions,
		},
		NotBefore:   time.Now().Add(-time.Hour),
		NotAfter:    time.Now().Add(24 * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageAny},
	}
	der, _ := x509.CreateCertificate(rand.Reader, template, ca.cert, pub, ca.key)

	certFile := filepath.Join(ca.dir, name+".pem")
	keyFile := filepath.Join(ca.dir, name+"-key.pem")
	writePEM(t, certFile, "CERTIFICATE", der)
	keyBytes, _ := x509.MarshalPKCS8PrivateKey(key)
	writePEM(t, keyFile, "PRIVATE KEY", keyBytes)

	return certFile, keyFile
}

func (ca *testCA) securityConfig() config.Security {
	return config.Security{
		CACertPath:  filepath.Join(ca.dir, "ca.pem"),
		CertPath:    filepath.Join(ca.dir, "server.pem"),
		KeyPath:     filepath.Join(ca.dir, "server-key.pem"),
		CryptoAlgo:  "ed25519",
		RequireMTLS: true,
	}
}

func writePEM(t *testing.T, path, blockType string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: blockType, Bytes: data}), 0600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// --- Server helpers ---

type Item struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ItemList struct {
	Items []Item `json:"items"`
}

type CreateItemRequest struct {
	Name string `json:"name"`
}

type DeleteResult struct {
	Deleted bool `json:"deleted"`
}

// startServer boots a rocco engine with mTLS, scoped test endpoints, and
// returns the base URL.
func startServer(t *testing.T, ca *testCA) string {
	t.Helper()

	sctx.ResetAdminForTesting()

	cfg := ca.securityConfig()
	authority, err := auth.New(context.Background(), cfg)
	if err != nil {
		t.Fatalf("auth.New: %v", err)
	}
	tlsCfg, err := auth.TLSConfig(cfg)
	if err != nil {
		t.Fatalf("auth.TLSConfig: %v", err)
	}

	engine := rocco.NewEngine().
		WithMiddleware(auth.Terminate(authority)).
		WithAuthenticator(auth.Authenticator())

	// Public: anyone authenticated can call.
	listItems := rocco.GET[rocco.NoBody, ItemList]("/items", func(_ *rocco.Request[rocco.NoBody]) (ItemList, error) {
		return ItemList{Items: []Item{{ID: "1", Name: "alpha"}, {ID: "2", Name: "beta"}}}, nil
	}).WithSummary("List items").WithAuthentication().WithScopes("read:items")

	// Write scope required.
	createItem := rocco.POST[CreateItemRequest, Item]("/items", func(r *rocco.Request[CreateItemRequest]) (Item, error) {
		return Item{ID: "3", Name: r.Body.Name}, nil
	}).WithSummary("Create item").WithAuthentication().WithScopes("write:items")

	// Admin scope required.
	deleteItem := rocco.DELETE[rocco.NoBody, DeleteResult]("/items/{id}", func(_ *rocco.Request[rocco.NoBody]) (DeleteResult, error) {
		return DeleteResult{Deleted: true}, nil
	}).WithSummary("Delete item").WithAuthentication().WithScopes("admin:items")

	engine.WithHandlers(listItems, createItem, deleteItem)

	ln, err := tls.Listen("tcp", "127.0.0.1:0", tlsCfg)
	if err != nil {
		t.Fatalf("tls.Listen: %v", err)
	}

	server := &http.Server{Handler: engine.Router()}
	go func() { _ = server.Serve(ln) }()
	t.Cleanup(func() { _ = server.Close() })

	return fmt.Sprintf("https://127.0.0.1:%d", ln.Addr().(*net.TCPAddr).Port)
}

// --- MCP helpers ---

func newMCPServer(t *testing.T, baseURL string, ca *testCA, certPath, keyPath string) *mcp.Server {
	t.Helper()
	client, err := mcp.NewClient(baseURL, filepath.Join(ca.dir, "ca.pem"), certPath, keyPath)
	if err != nil {
		t.Fatalf("mcp.NewClient: %v", err)
	}
	server := mcp.NewServer(client)
	if err := server.LoadTools(); err != nil {
		t.Fatalf("LoadTools: %v", err)
	}
	return server
}

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

func mcpToolCall(t *testing.T, s *mcp.Server, toolName string, args string, id int) (text string, isError bool) {
	t.Helper()
	resp := mcpRoundTrip(t, s, "tools/call", id,
		fmt.Sprintf(`{"name":"%s","arguments":%s}`, toolName, args))
	result, ok := resp["result"].(map[string]any)
	if !ok {
		// Protocol error.
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

// --- Tests ---

func TestEndToEnd(t *testing.T) {
	ca := newTestCA(t)
	ca.issueServer(t, "kuang-server", []string{"read:items", "write:items", "admin:items"})
	baseURL := startServer(t, ca)

	certPath, keyPath := ca.issueClient(t, "full-agent", []string{"read:items", "write:items"})
	server := newMCPServer(t, baseURL, ca, certPath, keyPath)

	t.Logf("loaded %d tools", server.ToolCount())
	if server.ToolCount() == 0 {
		t.Fatal("expected at least 1 tool")
	}

	// tools/list returns all tools (spec is unfiltered).
	names := mcpToolNames(t, server)
	t.Logf("tools: %v", names)
	if len(names) < 2 {
		t.Fatalf("expected at least 2 tools, got %d", len(names))
	}

	// Find tool names by method hint.
	var getTool, postTool string
	for _, n := range names {
		lower := strings.ToLower(n)
		if strings.Contains(lower, "get") {
			getTool = n
		}
		if strings.Contains(lower, "post") || strings.Contains(lower, "create") {
			postTool = n
		}
	}

	// Call list items.
	text, isErr := mcpToolCall(t, server, getTool, `{}`, 10)
	if isErr {
		t.Fatalf("list items failed: %s", text)
	}
	if !strings.Contains(text, "alpha") {
		t.Errorf("expected 'alpha' in response, got: %s", text)
	}

	// Call create item.
	text, isErr = mcpToolCall(t, server, postTool, `{"body":{"name":"gamma"}}`, 11)
	if isErr {
		t.Fatalf("create item failed: %s", text)
	}
	if !strings.Contains(text, "gamma") {
		t.Errorf("expected 'gamma' in response, got: %s", text)
	}
}

func TestReadOnlyAgentCannotWrite(t *testing.T) {
	ca := newTestCA(t)
	ca.issueServer(t, "kuang-server", []string{"read:items", "write:items", "admin:items"})
	baseURL := startServer(t, ca)

	// Agent with only read:items.
	certPath, keyPath := ca.issueClient(t, "reader-agent", []string{"read:items"})
	server := newMCPServer(t, baseURL, ca, certPath, keyPath)

	names := mcpToolNames(t, server)
	t.Logf("reader tools: %v", names)

	// Find the POST/create tool.
	var postTool string
	for _, n := range names {
		if strings.Contains(strings.ToLower(n), "post") || strings.Contains(strings.ToLower(n), "create") {
			postTool = n
			break
		}
	}
	if postTool == "" {
		t.Fatal("could not find POST tool")
	}

	// Calling write endpoint should fail — rocco enforces scopes.
	text, isErr := mcpToolCall(t, server, postTool, `{"body":{"name":"blocked"}}`, 20)
	if !isErr {
		t.Fatalf("expected error calling write endpoint with read-only cert, got: %s", text)
	}
	t.Logf("reader blocked from write: %s", text)
}

func TestReadOnlyAgentCanRead(t *testing.T) {
	ca := newTestCA(t)
	ca.issueServer(t, "kuang-server", []string{"read:items", "write:items", "admin:items"})
	baseURL := startServer(t, ca)

	certPath, keyPath := ca.issueClient(t, "reader-agent", []string{"read:items"})
	server := newMCPServer(t, baseURL, ca, certPath, keyPath)

	var getTool string
	for _, n := range mcpToolNames(t, server) {
		if strings.Contains(strings.ToLower(n), "get") {
			getTool = n
			break
		}
	}
	if getTool == "" {
		t.Fatal("could not find GET tool")
	}

	text, isErr := mcpToolCall(t, server, getTool, `{}`, 30)
	if isErr {
		t.Fatalf("read-only agent should be able to read: %s", text)
	}
	if !strings.Contains(text, "alpha") {
		t.Errorf("expected 'alpha', got: %s", text)
	}
}

func TestNoPermissionsAgentBlockedFromEverything(t *testing.T) {
	ca := newTestCA(t)
	ca.issueServer(t, "kuang-server", []string{"read:items", "write:items", "admin:items"})
	baseURL := startServer(t, ca)

	// Agent with zero permissions.
	certPath, keyPath := ca.issueClient(t, "empty-agent", nil)
	server := newMCPServer(t, baseURL, ca, certPath, keyPath)

	names := mcpToolNames(t, server)
	t.Logf("empty-agent tools: %v", names)

	// Try every tool — all should fail.
	for i, name := range names {
		text, isErr := mcpToolCall(t, server, name, `{}`, 40+i)
		if !isErr {
			t.Errorf("tool %q should have been denied, got: %s", name, text)
		}
	}
}

func TestWriterAgentCannotDelete(t *testing.T) {
	ca := newTestCA(t)
	ca.issueServer(t, "kuang-server", []string{"read:items", "write:items", "admin:items"})
	baseURL := startServer(t, ca)

	// Agent with read + write but not admin.
	certPath, keyPath := ca.issueClient(t, "writer-agent", []string{"read:items", "write:items"})
	server := newMCPServer(t, baseURL, ca, certPath, keyPath)

	var deleteTool string
	for _, n := range mcpToolNames(t, server) {
		if strings.Contains(strings.ToLower(n), "delete") {
			deleteTool = n
			break
		}
	}
	if deleteTool == "" {
		t.Fatal("could not find DELETE tool")
	}

	// Delete requires admin:items — writer doesn't have it.
	text, isErr := mcpToolCall(t, server, deleteTool, `{}`, 50)
	if !isErr {
		t.Fatalf("expected error calling delete without admin scope, got: %s", text)
	}
	t.Logf("writer blocked from delete: %s", text)
}

func TestAdminAgentCanDoEverything(t *testing.T) {
	ca := newTestCA(t)
	ca.issueServer(t, "kuang-server", []string{"read:items", "write:items", "admin:items"})
	baseURL := startServer(t, ca)

	// Agent with all permissions.
	certPath, keyPath := ca.issueClient(t, "admin-agent", []string{"read:items", "write:items", "admin:items"})
	server := newMCPServer(t, baseURL, ca, certPath, keyPath)

	names := mcpToolNames(t, server)
	t.Logf("admin tools: %v", names)

	// Every tool should succeed.
	for i, name := range names {
		args := `{}`
		if strings.Contains(strings.ToLower(name), "post") || strings.Contains(strings.ToLower(name), "create") {
			args = `{"body":{"name":"test"}}`
		}
		text, isErr := mcpToolCall(t, server, name, args, 60+i)
		if isErr {
			t.Errorf("admin should have access to %q, got error: %s", name, text)
		}
	}
}
