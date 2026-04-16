//go:build testing

package auth

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sctx"

	"github.com/zoobz-io/kuang/config"
)

// testPKI holds all generated certificates and keys for a test run.
type testPKI struct {
	caKey      ed25519.PrivateKey
	caCert     *x509.Certificate
	caCertPEM  []byte
	serverKey  ed25519.PrivateKey
	serverCert *x509.Certificate
	serverPEM  []byte
	serverKPEM []byte
	clientKey  ed25519.PrivateKey
	clientCert *x509.Certificate
	caPool     *x509.CertPool
}

func generatePKI(t *testing.T, clientCN string, clientOUs []string) *testPKI {
	t.Helper()

	caPub, caKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate CA key: %v", err)
	}
	caTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{Organization: []string{"Test CA"}},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, caPub, caKey)
	if err != nil {
		t.Fatalf("create CA cert: %v", err)
	}
	caCert, _ := x509.ParseCertificate(caDER)
	caCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER})

	serverPub, serverKey, _ := ed25519.GenerateKey(rand.Reader)
	serverTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "kuang-server"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1)},
	}
	serverDER, _ := x509.CreateCertificate(rand.Reader, serverTemplate, caCert, serverPub, caKey)
	serverCert, _ := x509.ParseCertificate(serverDER)
	serverCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: serverDER})
	serverKeyBytes, _ := x509.MarshalPKCS8PrivateKey(serverKey)
	serverKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: serverKeyBytes})

	clientPub, clientKey, _ := ed25519.GenerateKey(rand.Reader)
	clientTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(3),
		Subject: pkix.Name{
			CommonName:         clientCN,
			OrganizationalUnit: clientOUs,
		},
		NotBefore:   time.Now().Add(-time.Hour),
		NotAfter:    time.Now().Add(24 * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageAny},
	}
	clientDER, _ := x509.CreateCertificate(rand.Reader, clientTemplate, caCert, clientPub, caKey)
	clientCert, _ := x509.ParseCertificate(clientDER)

	pool := x509.NewCertPool()
	pool.AddCert(caCert)

	return &testPKI{
		caKey: caKey, caCert: caCert, caCertPEM: caCertPEM,
		serverKey: serverKey, serverCert: serverCert,
		serverPEM: serverCertPEM, serverKPEM: serverKeyPEM,
		clientKey: clientKey, clientCert: clientCert,
		caPool: pool,
	}
}

func writePKIFiles(t *testing.T, pki *testPKI) config.Security {
	t.Helper()
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "ca.pem"), pki.caCertPEM, 0644)
	os.WriteFile(filepath.Join(dir, "server.pem"), pki.serverPEM, 0644)
	os.WriteFile(filepath.Join(dir, "server-key.pem"), pki.serverKPEM, 0600)

	return config.Security{
		CACertPath:  filepath.Join(dir, "ca.pem"),
		CertPath:    filepath.Join(dir, "server.pem"),
		KeyPath:     filepath.Join(dir, "server-key.pem"),
		CryptoAlgo:  "ed25519",
		RequireMTLS: true,
	}
}

func setupAuthority(t *testing.T, pki *testPKI) *Authority {
	t.Helper()
	sctx.ResetAdminForTesting()
	cfg := writePKIFiles(t, pki)
	auth, err := New(context.Background(), cfg)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	return auth
}

func tlsState(cert *x509.Certificate) *tls.ConnectionState {
	return &tls.ConnectionState{PeerCertificates: []*x509.Certificate{cert}}
}

// --- admin.go tests ---

func TestNew(t *testing.T) {
	sctx.ResetAdminForTesting()
	pki := generatePKI(t, "agent-1", []string{"read:users"})
	cfg := writePKIFiles(t, pki)

	auth, err := New(context.Background(), cfg)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if auth.Admin == nil {
		t.Fatal("expected non-nil admin")
	}
	if auth.Token == "" {
		t.Fatal("expected non-empty server token")
	}
}

func TestNewBadCACert(t *testing.T) {
	sctx.ResetAdminForTesting()
	cfg := config.Security{
		CACertPath: "/nonexistent/ca.pem",
		CertPath:   "/nonexistent/server.pem",
		KeyPath:    "/nonexistent/key.pem",
	}
	_, err := New(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected error for missing CA cert")
	}
}

func TestNewBadServerKeypair(t *testing.T) {
	sctx.ResetAdminForTesting()
	pki := generatePKI(t, "agent-1", nil)
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "ca.pem"), pki.caCertPEM, 0644)

	cfg := config.Security{
		CACertPath: filepath.Join(dir, "ca.pem"),
		CertPath:   "/nonexistent/server.pem",
		KeyPath:    "/nonexistent/key.pem",
	}
	_, err := New(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected error for missing server keypair")
	}
}

func TestTLSConfigRequireMTLS(t *testing.T) {
	pki := generatePKI(t, "agent-1", nil)
	cfg := writePKIFiles(t, pki)
	cfg.RequireMTLS = true

	tlsCfg, err := TLSConfig(cfg)
	if err != nil {
		t.Fatalf("TLSConfig() error: %v", err)
	}
	if tlsCfg.ClientAuth != tls.RequireAndVerifyClientCert {
		t.Errorf("expected RequireAndVerifyClientCert, got %v", tlsCfg.ClientAuth)
	}
	if tlsCfg.MinVersion != tls.VersionTLS13 {
		t.Error("expected TLS 1.3 minimum")
	}
}

func TestTLSConfigOptionalMTLS(t *testing.T) {
	pki := generatePKI(t, "agent-1", nil)
	cfg := writePKIFiles(t, pki)
	cfg.RequireMTLS = false

	tlsCfg, err := TLSConfig(cfg)
	if err != nil {
		t.Fatalf("TLSConfig() error: %v", err)
	}
	if tlsCfg.ClientAuth != tls.VerifyClientCertIfGiven {
		t.Errorf("expected VerifyClientCertIfGiven, got %v", tlsCfg.ClientAuth)
	}
}

func TestServerPrivateKey(t *testing.T) {
	pki := generatePKI(t, "agent-1", nil)
	cfg := writePKIFiles(t, pki)

	key, err := ServerPrivateKey(cfg)
	if err != nil {
		t.Fatalf("ServerPrivateKey() error: %v", err)
	}
	if key == nil {
		t.Fatal("expected non-nil key")
	}
}

func TestAgentPolicy(t *testing.T) {
	policy := agentPolicy()
	pki := generatePKI(t, "my-agent", []string{"read:data", "write:data"})

	ctx, err := policy(pki.clientCert)
	if err != nil {
		t.Fatalf("policy error: %v", err)
	}
	if ctx.Metadata.Agent != "my-agent" {
		t.Errorf("expected agent 'my-agent', got %q", ctx.Metadata.Agent)
	}
	if len(ctx.Permissions) != 2 {
		t.Fatalf("expected 2 permissions, got %d", len(ctx.Permissions))
	}
	if ctx.Permissions[0] != "read:data" || ctx.Permissions[1] != "write:data" {
		t.Errorf("unexpected permissions: %v", ctx.Permissions)
	}
}

// --- middleware.go tests ---

func TestIdentity(t *testing.T) {
	id := &Identity{ctx: &sctx.Context[AgentMeta]{
		Metadata: AgentMeta{Agent: "bot-1"},
		CertificateInfo: sctx.CertificateInfo{
			CommonName: "bot-1",
		},
		Permissions: []string{"read:users", "write:users"},
	}}

	if id.ID() != "bot-1" {
		t.Errorf("ID() = %q, want bot-1", id.ID())
	}
	if id.TenantID() != "bot-1" {
		t.Errorf("TenantID() = %q, want bot-1", id.TenantID())
	}
	if id.Email() != "" {
		t.Errorf("Email() = %q, want empty", id.Email())
	}
	if len(id.Scopes()) != 2 {
		t.Errorf("Scopes() len = %d, want 2", len(id.Scopes()))
	}
	if id.Roles() != nil {
		t.Errorf("Roles() should be nil")
	}
	if !id.HasScope("read:users") {
		t.Error("HasScope(read:users) should be true")
	}
	if id.HasScope("delete:users") {
		t.Error("HasScope(delete:users) should be false")
	}
	if id.HasRole("admin") {
		t.Error("HasRole should always be false")
	}
	if id.Stats() != nil {
		t.Error("Stats() should be nil")
	}
}

func TestIdentityToken(t *testing.T) {
	id := &Identity{
		ctx:   &sctx.Context[AgentMeta]{},
		token: sctx.SignedToken("test-token"),
	}
	if id.Token() != "test-token" {
		t.Errorf("Token() = %q, want test-token", id.Token())
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	writeError(w, rocco.ErrUnauthorized.WithMessage("test message"))

	if w.Code != 401 {
		t.Errorf("status = %d, want 401", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("content-type = %q, want application/json", ct)
	}

	var body errBody
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Message != "test message" {
		t.Errorf("message = %q, want 'test message'", body.Message)
	}
}

// --- Terminate middleware tests ---

func TestTerminateRejectsNoTLS(t *testing.T) {
	pki := generatePKI(t, "agent-1", []string{"read:data"})
	authority := setupAuthority(t, pki)

	handler := Terminate(authority)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestTerminateRejectsEmptyPeerCerts(t *testing.T) {
	pki := generatePKI(t, "agent-1", []string{"read:data"})
	authority := setupAuthority(t, pki)

	handler := Terminate(authority)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.TLS = &tls.ConnectionState{}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestTerminateAcceptsValidCert(t *testing.T) {
	pki := generatePKI(t, "agent-1", []string{"read:data"})
	authority := setupAuthority(t, pki)

	var gotIdentity *Identity
	handler := Terminate(authority)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := r.Context().Value(identityKey{}).(*Identity)
		if !ok {
			t.Fatal("identity not in context")
		}
		gotIdentity = id
		w.WriteHeader(200)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.TLS = tlsState(pki.clientCert)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if gotIdentity == nil {
		t.Fatal("identity was not injected")
	}
	if gotIdentity.ID() != "agent-1" {
		t.Errorf("identity ID = %q, want agent-1", gotIdentity.ID())
	}
	if !gotIdentity.HasScope("read:data") {
		t.Error("identity should have read:data scope")
	}
	if gotIdentity.Token() == "" {
		t.Error("identity should carry a token")
	}
}

func TestTerminateCachesContext(t *testing.T) {
	pki := generatePKI(t, "agent-1", []string{"read:data"})
	authority := setupAuthority(t, pki)

	callCount := 0
	handler := Terminate(authority)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(200)
	}))

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.TLS = tlsState(pki.clientCert)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("request %d: status = %d, want 200", i, w.Code)
		}
	}

	if callCount != 3 {
		t.Errorf("handler called %d times, want 3", callCount)
	}
}

// --- IdentityFromContext tests ---

func TestIdentityFromContext(t *testing.T) {
	id := &Identity{ctx: &sctx.Context[AgentMeta]{
		Metadata:        AgentMeta{Agent: "bot"},
		CertificateInfo: sctx.CertificateInfo{CommonName: "bot"},
	}}
	ctx := context.WithValue(context.Background(), identityKey{}, id)

	got := IdentityFromContext(ctx)
	if got == nil {
		t.Fatal("expected non-nil identity")
	}
	if got.ID() != "bot" {
		t.Errorf("ID = %q, want bot", got.ID())
	}
}

func TestIdentityFromContextMissing(t *testing.T) {
	got := IdentityFromContext(context.Background())
	if got != nil {
		t.Error("expected nil identity from empty context")
	}
}

// --- Authenticator tests ---

func TestAuthenticatorWithIdentity(t *testing.T) {
	id := &Identity{ctx: &sctx.Context[AgentMeta]{
		Metadata:        AgentMeta{Agent: "bot"},
		CertificateInfo: sctx.CertificateInfo{CommonName: "bot"},
	}}

	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), identityKey{}, id)
	req = req.WithContext(ctx)

	result := Authenticator()(req)
	if result.ID() != "bot" {
		t.Errorf("ID = %q, want bot", result.ID())
	}
}

func TestAuthenticatorWithoutIdentity(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	result := Authenticator()(req)

	if _, ok := result.(rocco.NoIdentity); !ok {
		t.Errorf("expected NoIdentity, got %T", result)
	}
}
