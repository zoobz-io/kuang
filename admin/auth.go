package admin

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/zoobz-io/rocco"
)

// AdminIdentity implements rocco.Identity for the admin operator.
type AdminIdentity struct{}

func (i *AdminIdentity) ID() string            { return "admin" }
func (i *AdminIdentity) TenantID() string      { return "admin" }
func (i *AdminIdentity) Email() string          { return "" }
func (i *AdminIdentity) Scopes() []string       { return []string{"admin"} }
func (i *AdminIdentity) Roles() []string        { return nil }
func (i *AdminIdentity) HasScope(_ string) bool { return true }
func (i *AdminIdentity) HasRole(_ string) bool  { return false }
func (i *AdminIdentity) Stats() map[string]int  { return nil }

type adminIdentityKey struct{}

// Authenticate returns middleware that validates the API key from the
// Authorization header and injects an AdminIdentity into the request context.
func Authenticate(apiKey string) func(http.Handler) http.Handler {
	keyBytes := []byte(apiKey)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			token, ok := strings.CutPrefix(header, "Bearer ")
			if !ok || subtle.ConstantTimeCompare([]byte(token), keyBytes) != 1 {
				writeError(w, http.StatusUnauthorized, "unauthorized", "invalid or missing API key")
				return
			}
			ctx := context.WithValue(r.Context(), adminIdentityKey{}, &AdminIdentity{})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Authenticator returns a rocco identity extractor that reads the
// AdminIdentity injected by the Authenticate middleware.
func Authenticator() func(context.Context, *http.Request) (rocco.Identity, error) {
	return func(_ context.Context, r *http.Request) (rocco.Identity, error) {
		if id, ok := r.Context().Value(adminIdentityKey{}).(*AdminIdentity); ok {
			return id, nil
		}
		return rocco.NoIdentity{}, nil
	}
}

// writeError writes a JSON error response matching rocco's convention.
func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}{Code: code, Message: message})
}
