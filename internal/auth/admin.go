// Package auth provides mTLS termination and sctx-based authorization.
package auth

import (
	"context"
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/zoobz-io/kuang/config"
	"github.com/zoobz-io/sctx"
)

// AgentMeta is the metadata carried in every security context.
type AgentMeta struct {
	Agent string
}

// Authority holds the sctx Admin and the server's own token. The server
// token is required for creating guards — it defines the permission ceiling.
type Authority struct {
	Admin sctx.Admin[AgentMeta]
	Token sctx.SignedToken
}

// New bootstraps the sctx Admin from the security config. It loads the CA
// pool, server certificate, and private key from disk, then generates a
// trusted token for the server's own cert so it can create guards.
func New(ctx context.Context, cfg config.Security) (*Authority, error) {
	caPEM, err := os.ReadFile(cfg.CACertPath)
	if err != nil {
		return nil, fmt.Errorf("read CA cert: %w", err)
	}
	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caPEM) {
		return nil, fmt.Errorf("failed to parse CA cert")
	}

	serverTLS, err := tls.LoadX509KeyPair(cfg.CertPath, cfg.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("load server keypair: %w", err)
	}

	admin, err := sctx.NewAdminService[AgentMeta](serverTLS.PrivateKey, caPool)
	if err != nil {
		return nil, fmt.Errorf("create admin: %w", err)
	}

	if err = admin.SetPolicy(agentPolicy()); err != nil {
		return nil, fmt.Errorf("set policy: %w", err)
	}
	if err = admin.SetCache(sctx.NewBoundedMemoryCache[AgentMeta](sctx.CacheOptions{
		MaxSize: 256,
	})); err != nil {
		return nil, fmt.Errorf("set cache: %w", err)
	}

	// Parse the server's leaf certificate and generate a trusted token.
	// This token defines the permission ceiling for guard creation.
	serverCert, err := x509.ParseCertificate(serverTLS.Certificate[0])
	if err != nil {
		return nil, fmt.Errorf("parse server cert: %w", err)
	}
	token, err := admin.GenerateTrusted(ctx, serverCert)
	if err != nil {
		return nil, fmt.Errorf("generate server token: %w", err)
	}

	return &Authority{Admin: admin, Token: token}, nil
}

// TLSConfig returns a tls.Config configured for mTLS using the given security
// config. Clients must present a valid certificate signed by the trusted CA.
func TLSConfig(cfg config.Security) (*tls.Config, error) {
	caPEM, err := os.ReadFile(cfg.CACertPath)
	if err != nil {
		return nil, fmt.Errorf("read CA cert: %w", err)
	}
	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caPEM) {
		return nil, fmt.Errorf("failed to parse CA cert")
	}

	serverCert, err := tls.LoadX509KeyPair(cfg.CertPath, cfg.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("load server keypair: %w", err)
	}

	clientAuth := tls.RequireAndVerifyClientCert
	if !cfg.RequireMTLS {
		clientAuth = tls.VerifyClientCertIfGiven
	}

	return &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   clientAuth,
		ClientCAs:    caPool,
		MinVersion:   tls.VersionTLS13,
	}, nil
}

// ServerPrivateKey extracts the private key from the server certificate for
// use in assertion verification.
func ServerPrivateKey(cfg config.Security) (crypto.PrivateKey, error) {
	serverCert, err := tls.LoadX509KeyPair(cfg.CertPath, cfg.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("load server keypair: %w", err)
	}
	return serverCert.PrivateKey, nil
}

// agentPolicy maps certificate fields to security contexts. The agent's CN
// becomes the identity and OU entries become permissions.
func agentPolicy() sctx.ContextPolicy[AgentMeta] {
	return func(cert *x509.Certificate) (*sctx.Context[AgentMeta], error) {
		base, err := sctx.DefaultContextPolicy[AgentMeta]()(cert)
		if err != nil {
			return nil, err
		}
		base.Metadata = AgentMeta{Agent: cert.Subject.CommonName}

		// OU entries on the cert map directly to permissions.
		// e.g. OU=read:users,write:users grants those scopes.
		base.Permissions = cert.Subject.OrganizationalUnit

		return base, nil
	}
}
