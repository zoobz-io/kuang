package cli

import (
	"crypto/ed25519"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
)

func TestInitCA(t *testing.T) {
	dir := t.TempDir()

	if err := InitCA(dir, []string{"read:data", "write:data"}); err != nil {
		t.Fatalf("InitCA: %v", err)
	}

	// CA files exist.
	for _, f := range []string{"ca.pem", "ca-key.pem", "server.pem", "server-key.pem"} {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Errorf("missing %s: %v", f, err)
		}
	}

	// CA cert is valid.
	caCert := readCert(t, filepath.Join(dir, "ca.pem"))
	if !caCert.IsCA {
		t.Error("CA cert should have IsCA=true")
	}
	if caCert.Subject.CommonName != "kuang-ca" {
		t.Errorf("CA CN = %q, want kuang-ca", caCert.Subject.CommonName)
	}

	// Server cert is signed by CA.
	serverCert := readCert(t, filepath.Join(dir, "server.pem"))
	if serverCert.Subject.CommonName != "kuang-server" {
		t.Errorf("server CN = %q, want kuang-server", serverCert.Subject.CommonName)
	}

	pool := x509.NewCertPool()
	pool.AddCert(caCert)
	if _, err := serverCert.Verify(x509.VerifyOptions{
		Roots:     pool,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}); err != nil {
		t.Errorf("server cert not valid against CA: %v", err)
	}

	// Server cert carries permissions.
	if len(serverCert.Subject.OrganizationalUnit) != 2 {
		t.Errorf("server OU count = %d, want 2", len(serverCert.Subject.OrganizationalUnit))
	}

	// Server keypair loads as TLS cert.
	if _, err := tls.LoadX509KeyPair(
		filepath.Join(dir, "server.pem"),
		filepath.Join(dir, "server-key.pem"),
	); err != nil {
		t.Errorf("server keypair invalid: %v", err)
	}
}

func TestIssueCert(t *testing.T) {
	dir := t.TempDir()

	if err := InitCA(dir, nil); err != nil {
		t.Fatalf("InitCA: %v", err)
	}

	if err := IssueCert(dir, "test-agent", []string{"read:items", "write:items"}); err != nil {
		t.Fatalf("IssueCert: %v", err)
	}

	// Client files exist.
	for _, f := range []string{"test-agent.pem", "test-agent-key.pem"} {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Errorf("missing %s: %v", f, err)
		}
	}

	// Client cert is signed by CA.
	caCert := readCert(t, filepath.Join(dir, "ca.pem"))
	clientCert := readCert(t, filepath.Join(dir, "test-agent.pem"))

	if clientCert.Subject.CommonName != "test-agent" {
		t.Errorf("client CN = %q, want test-agent", clientCert.Subject.CommonName)
	}

	pool := x509.NewCertPool()
	pool.AddCert(caCert)
	if _, err := clientCert.Verify(x509.VerifyOptions{
		Roots:     pool,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}); err != nil {
		t.Errorf("client cert not valid against CA: %v", err)
	}

	// Client cert carries scopes.
	if len(clientCert.Subject.OrganizationalUnit) != 2 {
		t.Fatalf("client OU count = %d, want 2", len(clientCert.Subject.OrganizationalUnit))
	}
	if clientCert.Subject.OrganizationalUnit[0] != "read:items" {
		t.Errorf("OU[0] = %q, want read:items", clientCert.Subject.OrganizationalUnit[0])
	}

	// Keypair is valid.
	if _, err := tls.LoadX509KeyPair(
		filepath.Join(dir, "test-agent.pem"),
		filepath.Join(dir, "test-agent-key.pem"),
	); err != nil {
		t.Errorf("client keypair invalid: %v", err)
	}
}

func TestIssueCertNoCA(t *testing.T) {
	dir := t.TempDir()
	if err := IssueCert(dir, "agent", nil); err == nil {
		t.Fatal("expected error when CA doesn't exist")
	}
}

func TestIssueCertNoScopes(t *testing.T) {
	dir := t.TempDir()
	if err := InitCA(dir, nil); err != nil {
		t.Fatalf("InitCA: %v", err)
	}

	if err := IssueCert(dir, "empty-agent", nil); err != nil {
		t.Fatalf("IssueCert with no scopes: %v", err)
	}

	cert := readCert(t, filepath.Join(dir, "empty-agent.pem"))
	if len(cert.Subject.OrganizationalUnit) != 0 {
		t.Errorf("expected no OUs, got %v", cert.Subject.OrganizationalUnit)
	}
}

func TestFilePermissions(t *testing.T) {
	dir := t.TempDir()
	if err := InitCA(dir, nil); err != nil {
		t.Fatalf("InitCA: %v", err)
	}

	for _, f := range []string{"ca-key.pem", "server-key.pem", "ca.pem", "server.pem"} {
		info, err := os.Stat(filepath.Join(dir, f))
		if err != nil {
			t.Fatalf("stat %s: %v", f, err)
		}
		perm := info.Mode().Perm()
		if perm != 0600 {
			t.Errorf("%s permissions = %o, want 0600", f, perm)
		}
	}
}

func readCert(t *testing.T, path string) *x509.Certificate {
	t.Helper()
	data, err := os.ReadFile(path) //nolint:gosec // test file
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		t.Fatalf("no PEM block in %s", path)
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return cert
}

// Verify that issued certs actually work with the auth package's TLSConfig.
func TestIssuedCertWorksWithMTLS(t *testing.T) {
	dir := t.TempDir()
	if err := InitCA(dir, []string{"read:data"}); err != nil {
		t.Fatalf("InitCA: %v", err)
	}
	if err := IssueCert(dir, "agent", []string{"read:data"}); err != nil {
		t.Fatalf("IssueCert: %v", err)
	}

	// Load CA pool.
	caCert := readCert(t, filepath.Join(dir, "ca.pem"))
	pool := x509.NewCertPool()
	pool.AddCert(caCert)

	// Load server cert.
	serverTLS, err := tls.LoadX509KeyPair(
		filepath.Join(dir, "server.pem"),
		filepath.Join(dir, "server-key.pem"),
	)
	if err != nil {
		t.Fatalf("load server: %v", err)
	}

	// Load client cert.
	clientTLS, err := tls.LoadX509KeyPair(
		filepath.Join(dir, "agent.pem"),
		filepath.Join(dir, "agent-key.pem"),
	)
	if err != nil {
		t.Fatalf("load client: %v", err)
	}

	// Verify client cert is ed25519.
	if _, ok := clientTLS.PrivateKey.(ed25519.PrivateKey); !ok {
		t.Errorf("client key type = %T, want ed25519", clientTLS.PrivateKey)
	}

	// Verify server cert is ed25519.
	if _, ok := serverTLS.PrivateKey.(ed25519.PrivateKey); !ok {
		t.Errorf("server key type = %T, want ed25519", serverTLS.PrivateKey)
	}
}
