// Package admin provides an operator-facing HTTP API for managing kuang
// credentials and (eventually) agent lifecycle. It shares the same SQLite
// database as the main kuang server but uses API key authentication
// instead of mTLS.
package admin

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
	"github.com/zoobz-io/kuang/internal/creds"
	"github.com/zoobz-io/sum"
)

// Run bootstraps the admin server. It loads configuration, opens the shared
// credential store, registers endpoints, and starts a plain HTTP server
// with graceful shutdown.
func Run() error {
	var cfg Config
	if err := fig.Load(&cfg); err != nil {
		return fmt.Errorf("load admin config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("validate admin config: %w", err)
	}

	store, err := creds.Open(cfg.DBPath)
	if err != nil {
		return fmt.Errorf("open credential store: %w", err)
	}
	defer func() { _ = store.Close() }()

	svc := sum.New()
	k := sum.Start()
	sum.Freeze(k)

	svc.Engine().WithAuthenticator(Authenticator())
	svc.Handle(endpoints(store)...)

	var handler http.Handler = svc.Engine().Router()
	handler = Authenticate(cfg.APIKey)(handler)

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	server := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("admin: listening on %s", addr)
		errCh <- server.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case <-sigCh:
		log.Println("admin: shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return server.Shutdown(ctx)
	}
}

// NewHandler builds the admin HTTP handler for testing. It wires up the
// rocco engine, credential endpoints, and API key middleware without
// starting a server.
func NewHandler(store *creds.Store, apiKey string) http.Handler {
	svc := sum.New()
	k := sum.Start()
	sum.Freeze(k)

	svc.Engine().WithAuthenticator(Authenticator())
	svc.Handle(endpoints(store)...)

	var handler http.Handler = svc.Engine().Router()
	handler = Authenticate(apiKey)(handler)
	return handler
}
