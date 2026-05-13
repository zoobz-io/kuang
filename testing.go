//go:build testing

package kuang

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sctx"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/kuang/config"
	"github.com/zoobz-io/kuang/internal/auth"
)

// TestEnv is a running kuang server for integration tests. It manages the
// full PKI lifecycle and provides helpers for issuing client certificates
// with specific scopes.
type TestEnv struct {
	// BaseURL is the HTTPS URL of the running server (e.g. "https://127.0.0.1:9123").
	BaseURL string

	// CACertPath is the path to the CA certificate for MCP client connections.
	CACertPath string

	dir    string
	caKey  ed25519.PrivateKey
	caCert *x509.Certificate
	serial int64
}

// NewTestEnv boots a kuang server with the given modules for integration
// testing. It generates a test CA, server certificate, wires up mTLS + sctx,
// and starts serving on a random port. The server shuts down when the test
// completes.
//
// scopes defines the server's permission ceiling — the full set of scopes
// that agents can be granted. Pass the union of all module scopes.
func NewTestEnv(t *testing.T, scopes []string, modules ...Module) *TestEnv {
	t.Helper()

	// Reset singletons for test isolation.
	sctx.ResetAdminForTesting()
	sum.Reset()

	dir := t.TempDir()
	env := &TestEnv{dir: dir, serial: 1}

	// Generate CA.
	caPub, caKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate CA key: %v", err)
	}
	env.caKey = caKey

	caTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{Organization: []string{"kuang-test"}},
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
	env.caCert, _ = x509.ParseCertificate(caDER)
	env.CACertPath = filepath.Join(dir, "ca.pem")
	writePEMFile(t, env.CACertPath, "CERTIFICATE", caDER)

	// Generate server cert with all scopes as permission ceiling.
	env.serial++
	serverPub, serverKey, _ := ed25519.GenerateKey(rand.Reader)
	serverTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(env.serial),
		Subject: pkix.Name{
			CommonName:         "kuang-test-server",
			OrganizationalUnit: scopes,
		},
		NotBefore:   time.Now().Add(-time.Hour),
		NotAfter:    time.Now().Add(24 * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageAny},
		DNSNames:    []string{"localhost"},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1)},
	}
	serverDER, _ := x509.CreateCertificate(rand.Reader, serverTemplate, env.caCert, serverPub, caKey)
	writePEMFile(t, filepath.Join(dir, "server.pem"), "CERTIFICATE", serverDER)
	serverKeyBytes, _ := x509.MarshalPKCS8PrivateKey(serverKey)
	writePEMFile(t, filepath.Join(dir, "server-key.pem"), "PRIVATE KEY", serverKeyBytes)

	// Security config.
	secCfg := config.Security{
		CACertPath:  env.CACertPath,
		CertPath:    filepath.Join(dir, "server.pem"),
		KeyPath:     filepath.Join(dir, "server-key.pem"),
		CryptoAlgo:  "ed25519",
		RequireMTLS: true,
	}

	ctx := context.Background()

	authority, err := auth.New(ctx, secCfg)
	if err != nil {
		t.Fatalf("auth.New: %v", err)
	}
	tlsCfg, err := auth.TLSConfig(secCfg)
	if err != nil {
		t.Fatalf("auth.TLSConfig: %v", err)
	}

	// Sum + rocco setup.
	svc := sum.New()
	k := sum.Start()

	var endpoints []rocco.Endpoint
	for _, mod := range modules {
		eps, err := mod(ctx, k)
		if err != nil {
			t.Fatalf("module: %v", err)
		}
		endpoints = append(endpoints, eps...)
	}

	svc.Engine().WithAuthenticator(auth.Authenticator())
	svc.Handle(endpoints...)
	sum.Freeze(k)

	// Start TLS server on random port.
	var handler http.Handler = openAPIInterceptor(svc.Engine())
	handler = auth.Terminate(authority)(handler)

	ln, err := tls.Listen("tcp", "127.0.0.1:0", tlsCfg)
	if err != nil {
		t.Fatalf("tls.Listen: %v", err)
	}

	server := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() { _ = server.Serve(ln) }()
	t.Cleanup(func() { _ = server.Close() })

	env.BaseURL = fmt.Sprintf("https://127.0.0.1:%d", ln.Addr().(*net.TCPAddr).Port)
	return env
}

// IssueClient generates a client certificate signed by the test CA with the
// given name and scopes. Returns paths to the cert and key files.
func (e *TestEnv) IssueClient(t *testing.T, name string, scopes []string) (certPath, keyPath string) {
	t.Helper()

	e.serial++
	pub, key, _ := ed25519.GenerateKey(rand.Reader)
	template := &x509.Certificate{
		SerialNumber: big.NewInt(e.serial),
		Subject: pkix.Name{
			CommonName:         name,
			OrganizationalUnit: scopes,
		},
		NotBefore:   time.Now().Add(-time.Hour),
		NotAfter:    time.Now().Add(24 * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageAny},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, e.caCert, pub, e.caKey)
	if err != nil {
		t.Fatalf("create client cert: %v", err)
	}

	certPath = filepath.Join(e.dir, name+".pem")
	keyPath = filepath.Join(e.dir, name+"-key.pem")
	writePEMFile(t, certPath, "CERTIFICATE", der)
	keyBytes, _ := x509.MarshalPKCS8PrivateKey(key)
	writePEMFile(t, keyPath, "PRIVATE KEY", keyBytes)

	return certPath, keyPath
}

func writePEMFile(t *testing.T, path, blockType string, data []byte) {
	t.Helper()
	err := os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: blockType, Bytes: data}), 0600)
	if err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
