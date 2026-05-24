// Package admin provides an operator-facing HTTP API for managing kuang
// credentials and (eventually) agent lifecycle. It shares the same SQLite
// database as the main kuang server. Authentication is pluggable via a
// rocco-compatible authenticator function — kuang ships with a local
// username/password implementation, but users can supply their own.
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
	"github.com/zoobz-io/kuang/internal/auth"
	"github.com/zoobz-io/kuang/internal/creds"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

// Authenticator is the function signature for admin authentication.
// It is the same signature that rocco.Engine.WithAuthenticator accepts.
type Authenticator = func(context.Context, *http.Request) (rocco.Identity, error)

// Option configures the admin server.
type Option func(*options)

type options struct {
	authenticator Authenticator
}

// WithAuthenticator sets a custom authenticator for the admin API.
// When not set, the server uses LocalAuthenticator backed by the
// configured database.
func WithAuthenticator(auth Authenticator) Option {
	return func(o *options) {
		o.authenticator = auth
	}
}

// Run bootstraps the admin server. It loads configuration, opens the shared
// credential store, registers endpoints, and starts a plain HTTP server
// with graceful shutdown.
func Run(opts ...Option) error {
	var cfg Config
	if err := fig.Load(&cfg); err != nil {
		return fmt.Errorf("load admin config: %w", err)
	}

	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	store, err := creds.Open(cfg.DBPath)
	if err != nil {
		return fmt.Errorf("open credential store: %w", err)
	}
	defer func() { _ = store.Close() }()

	authenticator := o.authenticator
	var userStore *auth.Store
	if authenticator == nil {
		userStore, err = auth.OpenStore(cfg.DBPath)
		if err != nil {
			return fmt.Errorf("open user store: %w", err)
		}
		defer func() { _ = userStore.Close() }()
		authenticator = LocalAuthenticator(userStore)
	}

	handler := newHandler(store, authenticator, userStore)

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

// NewHandler builds the admin HTTP handler for testing or embedding.
// It wires up the rocco engine, credential endpoints, and the given
// authenticator without starting a server. Pass a non-nil userStore
// to enable the first-time setup endpoint (local auth only).
func NewHandler(store *creds.Store, authenticator Authenticator, userStore *auth.Store) http.Handler {
	return newHandler(store, authenticator, userStore)
}

func newHandler(store *creds.Store, authenticator Authenticator, userStore *auth.Store) http.Handler {
	svc := sum.New()
	k := sum.Start()
	sum.Freeze(k)

	svc.Engine().WithAuthenticator(authenticator)

	var setupEndpoint []rocco.Endpoint
	if userStore != nil {
		setupEndpoint = append(setupEndpoint, setup(userStore))
	}
	svc.Handle(endpoints(store, setupEndpoint...)...)

	return svc.Engine().Router()
}
