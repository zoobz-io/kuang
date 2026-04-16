// Package cli implements the kuang CLI commands for managing servers and certificates.
package cli

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

// DefaultCertDir is the default directory for storing certificates.
const DefaultCertDir = "certs"

// InitCA generates a new certificate authority and server cert in the given
// directory. It writes ca.pem, ca-key.pem, server.pem, and server-key.pem.
// The server cert includes localhost and 127.0.0.1 for local development.
func InitCA(dir string, serverPermissions []string) error {
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("create cert dir: %w", err)
	}

	// Generate CA.
	caPub, caKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("generate CA key: %w", err)
	}

	caTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"kuang"},
			CommonName:   "kuang-ca",
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}

	caDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, caPub, caKey)
	if err != nil {
		return fmt.Errorf("create CA cert: %w", err)
	}

	caCert, err := x509.ParseCertificate(caDER)
	if err != nil {
		return fmt.Errorf("parse CA cert: %w", err)
	}

	if err = writeCert(filepath.Join(dir, "ca.pem"), caDER); err != nil {
		return err
	}
	if err = writeKey(filepath.Join(dir, "ca-key.pem"), caKey); err != nil {
		return err
	}

	// Generate server cert signed by CA.
	serverPub, serverKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("generate server key: %w", err)
	}

	serverTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			Organization:       []string{"kuang"},
			CommonName:         "kuang-server",
			OrganizationalUnit: serverPermissions,
		},
		NotBefore:   time.Now().Add(-time.Hour),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageAny},
		DNSNames:    []string{"localhost"},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1)},
	}

	serverDER, err := x509.CreateCertificate(rand.Reader, serverTemplate, caCert, serverPub, caKey)
	if err != nil {
		return fmt.Errorf("create server cert: %w", err)
	}

	if err := writeCert(filepath.Join(dir, "server.pem"), serverDER); err != nil {
		return err
	}
	return writeKey(filepath.Join(dir, "server-key.pem"), serverKey)
}

// IssueCert generates a client certificate signed by the CA in the given
// directory. It writes {name}.pem and {name}-key.pem.
func IssueCert(dir, name string, scopes []string) error {
	caKey, caCert, err := loadCA(dir)
	if err != nil {
		return fmt.Errorf("load CA: %w", err)
	}

	serial, err := randomSerial()
	if err != nil {
		return err
	}

	clientPub, clientKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("generate client key: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization:       []string{"kuang"},
			CommonName:         name,
			OrganizationalUnit: scopes,
		},
		NotBefore:   time.Now().Add(-time.Hour),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageAny},
	}

	clientDER, err := x509.CreateCertificate(rand.Reader, template, caCert, clientPub, caKey)
	if err != nil {
		return fmt.Errorf("create client cert: %w", err)
	}

	if err := writeCert(filepath.Join(dir, name+".pem"), clientDER); err != nil {
		return err
	}
	return writeKey(filepath.Join(dir, name+"-key.pem"), clientKey)
}

func loadCA(dir string) (ed25519.PrivateKey, *x509.Certificate, error) {
	certPEM, err := os.ReadFile(filepath.Join(dir, "ca.pem")) //nolint:gosec // user-configured cert dir
	if err != nil {
		return nil, nil, fmt.Errorf("read CA cert: %w", err)
	}
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, nil, fmt.Errorf("failed to decode CA cert PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("parse CA cert: %w", err)
	}

	keyPEM, err := os.ReadFile(filepath.Join(dir, "ca-key.pem")) //nolint:gosec // user-configured cert dir
	if err != nil {
		return nil, nil, fmt.Errorf("read CA key: %w", err)
	}
	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return nil, nil, fmt.Errorf("failed to decode CA key PEM")
	}
	rawKey, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("parse CA key: %w", err)
	}
	edKey, ok := rawKey.(ed25519.PrivateKey)
	if !ok {
		return nil, nil, fmt.Errorf("CA key is not ed25519")
	}

	return edKey, cert, nil
}

func randomSerial() (*big.Int, error) {
	limit := new(big.Int).Lsh(big.NewInt(1), 128)
	serial, err := rand.Int(rand.Reader, limit)
	if err != nil {
		return nil, fmt.Errorf("generate serial: %w", err)
	}
	return serial, nil
}

func writeCert(path string, der []byte) error {
	data := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	return os.WriteFile(path, data, 0600)
}

func writeKey(path string, key ed25519.PrivateKey) error {
	keyBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return fmt.Errorf("marshal key: %w", err)
	}
	return os.WriteFile(path, pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: keyBytes,
	}), 0600)
}
