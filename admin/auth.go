package admin

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/zoobz-io/kuang/internal/auth"
	"github.com/zoobz-io/rocco"
)

// AdminIdentity implements rocco.Identity for authenticated admin users.
type AdminIdentity struct {
	username string
	scopes   []string
}

func (i *AdminIdentity) ID() string       { return i.username }
func (i *AdminIdentity) TenantID() string { return "admin" }
func (i *AdminIdentity) Email() string    { return "" }
func (i *AdminIdentity) Scopes() []string { return i.scopes }
func (i *AdminIdentity) Roles() []string  { return nil }
func (i *AdminIdentity) HasScope(scope string) bool {
	for _, s := range i.scopes {
		if s == scope {
			return true
		}
	}
	return false
}
func (i *AdminIdentity) HasRole(_ string) bool { return false }
func (i *AdminIdentity) Stats() map[string]int { return nil }

// LocalAuthenticator returns a rocco-compatible authenticator backed by the
// given user store. It extracts Basic auth credentials from the request and
// verifies them against the store. On success it returns an AdminIdentity
// with the "admin" scope; on failure it returns rocco.NoIdentity so that
// rocco's scope enforcement rejects the request.
func LocalAuthenticator(store *auth.Store) func(context.Context, *http.Request) (rocco.Identity, error) {
	return func(ctx context.Context, r *http.Request) (rocco.Identity, error) {
		username, password, ok := parseBasicAuth(r.Header.Get("Authorization"))
		if !ok {
			return rocco.NoIdentity{}, nil
		}
		if err := store.Verify(ctx, username, password); err != nil {
			return rocco.NoIdentity{}, nil
		}
		return &AdminIdentity{
			username: username,
			scopes:   []string{ScopeAdmin},
		}, nil
	}
}

// parseBasicAuth extracts username and password from a Basic auth header value.
func parseBasicAuth(header string) (string, string, bool) {
	encoded, ok := strings.CutPrefix(header, "Basic ")
	if !ok {
		return "", "", false
	}
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", "", false
	}
	username, password, ok := strings.Cut(string(decoded), ":")
	return username, password, ok
}
