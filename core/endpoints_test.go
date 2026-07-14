//go:build testing

package core

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobzio/kuang/api/contracts"
	"github.com/zoobzio/kuang/internal/httpc"
)

// ---------------------------------------------------------------------------
// ResolveBearer failure branches.
//
// ResolveBearer first resolves a contracts.Credentials service from the sum
// registry, then reads the caller's identity from the context. Package core
// can reach both failure branches:
//
//   - store missing:    nothing registered -> rocco.ErrForbidden.
//   - identity missing: a store is registered but the context carries no
//                       auth.Identity -> rocco.ErrUnauthorized. Because
//                       auth.IdentityFromContext reads an unexported context
//                       key, package core cannot inject an identity, so a plain
//                       context always yields nil identity here.
//
// Seeding the registry uses the testing-only sum.Reset/Start/Register helpers,
// so this whole file is built under the "testing" tag.
// ---------------------------------------------------------------------------

func TestResolveBearerStoreMissing(t *testing.T) {
	opt, err := ResolveBearer(context.Background(), "github")

	if opt != nil {
		t.Errorf("opt = %v, want nil", opt)
	}
	if err == nil {
		t.Fatal("err = nil, want forbidden error")
	}
	if !errors.Is(err, rocco.ErrForbidden) {
		t.Errorf("errors.Is(err, ErrForbidden) = false, want true (err=%v)", err)
	}

	var def rocco.ErrorDefinition
	if !errors.As(err, &def) {
		t.Fatalf("err type = %T, want rocco.ErrorDefinition", err)
	}
	if got, want := def.Code(), "FORBIDDEN"; got != want {
		t.Errorf("Code() = %q, want %q", got, want)
	}
	if got, want := def.Status(), http.StatusForbidden; got != want {
		t.Errorf("Status() = %d, want %d", got, want)
	}
	if got, want := def.Message(), "credential store unavailable"; got != want {
		t.Errorf("Message() = %q, want %q", got, want)
	}
}

// fakeCredentials is a no-op contracts.Credentials used only to populate the
// sum registry so ResolveBearer gets past the store lookup and reaches the
// identity check. Its methods are never expected to run in these tests.
type fakeCredentials struct{}

func (fakeCredentials) Resolve(context.Context, string, string) (string, error) {
	return "", nil
}
func (fakeCredentials) Set(context.Context, string, string, string) error { return nil }
func (fakeCredentials) Delete(context.Context, string, string) error      { return nil }
func (fakeCredentials) List(context.Context, string) ([]string, error)    { return nil, nil }

// TestResolveBearerNoIdentity exercises the branch where the credential store
// is registered but no auth.Identity is present in the context. It seeds the
// process-global sum registry via the testing-only helpers and tears it down
// again so the empty-registry assertion in TestResolveBearerStoreMissing
// remains valid regardless of run order.
func TestResolveBearerNoIdentity(t *testing.T) {
	sum.Reset()
	t.Cleanup(sum.Reset)

	k := sum.Start()
	sum.Register[contracts.Credentials](k, fakeCredentials{})

	opt, err := ResolveBearer(context.Background(), "github")

	if opt != nil {
		t.Errorf("opt = %v, want nil", opt)
	}
	if err == nil {
		t.Fatal("err = nil, want unauthorized error")
	}
	if !errors.Is(err, rocco.ErrUnauthorized) {
		t.Errorf("errors.Is(err, ErrUnauthorized) = false, want true (err=%v)", err)
	}

	var def rocco.ErrorDefinition
	if !errors.As(err, &def) {
		t.Fatalf("err type = %T, want rocco.ErrorDefinition", err)
	}
	if got, want := def.Code(), "UNAUTHORIZED"; got != want {
		t.Errorf("Code() = %q, want %q", got, want)
	}
	if got, want := def.Status(), http.StatusUnauthorized; got != want {
		t.Errorf("Status() = %d, want %d", got, want)
	}
	if got, want := def.Message(), "identity required"; got != want {
		t.Errorf("Message() = %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// GET/POST/PUT/DELETE constructors.
//
// Each wraps the corresponding rocco constructor, so the returned handler must
// be non-nil and carry the expected method/path in its (public) Spec(). We do
// not invoke the handler here — doing so would run ResolveBearer, and its
// success path needs a live sum runtime plus an injected identity.
// ---------------------------------------------------------------------------

type ep struct{}

func noopHandler(*rocco.Request[ep], httpc.RequestOption) (ep, error) {
	return ep{}, nil
}

func TestEndpointConstructors(t *testing.T) {
	const (
		path    = "/widgets/{id}"
		credKey = "github"
	)

	tests := []struct {
		name       string
		build      func() (method string, path string, nonNil bool)
		wantMethod string
	}{
		{
			name: "GET",
			build: func() (string, string, bool) {
				h := GET[ep, ep](path, credKey, noopHandler)
				return h.Spec().Method, h.Spec().Path, h != nil
			},
			wantMethod: http.MethodGet,
		},
		{
			name: "POST",
			build: func() (string, string, bool) {
				h := POST[ep, ep](path, credKey, noopHandler)
				return h.Spec().Method, h.Spec().Path, h != nil
			},
			wantMethod: http.MethodPost,
		},
		{
			name: "PUT",
			build: func() (string, string, bool) {
				h := PUT[ep, ep](path, credKey, noopHandler)
				return h.Spec().Method, h.Spec().Path, h != nil
			},
			wantMethod: http.MethodPut,
		},
		{
			name: "DELETE",
			build: func() (string, string, bool) {
				h := DELETE[ep, ep](path, credKey, noopHandler)
				return h.Spec().Method, h.Spec().Path, h != nil
			},
			wantMethod: http.MethodDelete,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method, gotPath, nonNil := tt.build()
			if !nonNil {
				t.Fatal("handler = nil, want non-nil *rocco.Handler")
			}
			if method != tt.wantMethod {
				t.Errorf("Spec().Method = %q, want %q", method, tt.wantMethod)
			}
			if gotPath != path {
				t.Errorf("Spec().Path = %q, want %q", gotPath, path)
			}
		})
	}
}
