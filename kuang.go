// Package kuang provides a framework for building mTLS-secured API servers
// that expose tools to AI agents via MCP.
package kuang

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zoobz-io/fig"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/kuang/config"
	"github.com/zoobz-io/kuang/contracts"
	"github.com/zoobz-io/kuang/internal/auth"
	"github.com/zoobz-io/kuang/internal/creds"
)

// Module is a function that registers services and returns endpoints.
// Each module receives the application context and the sum registry key,
// allowing it to register services and load configuration before returning
// its HTTP endpoints.
type Module func(ctx context.Context, k sum.Key) ([]rocco.Endpoint, error)

// Run bootstraps the kuang server with the given modules. It loads security
// configuration from the environment, sets up mTLS, calls each module to
// collect endpoints, and starts the HTTPS server with graceful shutdown.
func Run(modules ...Module) error {
	ctx := context.Background()

	// Load configuration.
	var secCfg config.Security
	if err := fig.Load(&secCfg); err != nil {
		return fmt.Errorf("load security config: %w", err)
	}
	var srvCfg serverConfig
	if err := fig.Load(&srvCfg); err != nil {
		return fmt.Errorf("load server config: %w", err)
	}

	// Bootstrap sctx authority from certificates.
	authority, err := auth.New(ctx, secCfg)
	if err != nil {
		return fmt.Errorf("bootstrap auth: %w", err)
	}

	// Build TLS config for mTLS.
	tlsCfg, err := auth.TLSConfig(secCfg)
	if err != nil {
		return fmt.Errorf("tls config: %w", err)
	}

	// Open internal credential store.
	store, err := creds.Open(srvCfg.DBPath)
	if err != nil {
		return fmt.Errorf("open credential store: %w", err)
	}
	defer func() { _ = store.Close() }()

	// Initialize sum service and registry.
	svc := sum.New()
	k := sum.Start()

	// Register default credential store. Modules may override by
	// registering their own contracts.Credentials — slush replaces
	// silently on duplicate registration.
	sum.Register[contracts.Credentials](k, store)
	endpoints := []rocco.Endpoint{credsEndpoint()}

	// Run each module to collect endpoints.
	for _, mod := range modules {
		eps, err := mod(ctx, k)
		if err != nil {
			return fmt.Errorf("module: %w", err)
		}
		endpoints = append(endpoints, eps...)
	}

	// Configure the rocco engine with certificate-based authentication.
	svc.Engine().WithAuthenticator(auth.Authenticator())
	svc.Handle(endpoints...)
	sum.Freeze(k)

	// Build the handler chain. The openAPI interceptor serves a filtered
	// spec based on the requesting agent's scopes (rocco's built-in
	// /openapi serves an unfiltered spec). Terminate wraps everything to
	// extract client certificates and inject identity into the context.
	handler := openAPIInterceptor(svc.Engine())
	handler = auth.Terminate(authority)(handler)

	addr := fmt.Sprintf("%s:%d", srvCfg.Host, srvCfg.Port)
	server := &http.Server{
		Addr:              addr,
		Handler:           handler,
		TLSConfig:         tlsCfg,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Start server with graceful shutdown on SIGINT/SIGTERM.
	errCh := make(chan error, 1)
	go func() {
		log.Printf("kuang: listening on %s", addr)
		// Empty cert/key paths — TLS config already has the certificates.
		errCh <- server.ListenAndServeTLS("", "")
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case <-sigCh:
		log.Println("kuang: shutting down...")
		shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	}
}

// openAPIInterceptor wraps the rocco router to serve identity-filtered
// OpenAPI specs at GET /openapi. Rocco's built-in /openapi handler serves
// an unfiltered spec; this interceptor calls GenerateOpenAPI(identity) to
// filter endpoints by the requesting agent's scopes.
func openAPIInterceptor(engine *rocco.Engine) http.Handler {
	router := engine.Router()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/openapi" {
			var id rocco.Identity
			if identity := auth.IdentityFromContext(r.Context()); identity != nil {
				id = identity
			}
			spec := engine.GenerateOpenAPI(id)
			data, err := spec.ToJSON()
			if err != nil {
				http.Error(w, "failed to generate spec", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(data)
			return
		}
		router.ServeHTTP(w, r)
	})
}

type serverConfig struct {
	Host   string `env:"KUANG_HOST" default:"localhost"`
	DBPath string `env:"KUANG_DB_PATH" default:"data/kuang.db"`
	Port   int    `env:"KUANG_PORT" default:"8080"`
}
