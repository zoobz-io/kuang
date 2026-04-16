package auth

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sctx"
)

// errBody is the JSON error response format matching rocco's convention.
type errBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// writeError writes a rocco ErrorDefinition as a JSON response.
func writeError(w http.ResponseWriter, err rocco.ErrorDefinition) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Status())
	if encErr := json.NewEncoder(w).Encode(errBody{
		Code:    err.Code(),
		Message: err.Message(),
	}); encErr != nil {
		http.Error(w, err.Message(), err.Status())
	}
}

// Identity implements rocco.Identity backed by an sctx security context.
type Identity struct {
	ctx   *sctx.Context[AgentMeta]
	token sctx.SignedToken
}

// ID returns the agent's common name from the certificate.
func (i *Identity) ID() string { return i.ctx.CertificateInfo.CommonName }

// TenantID returns the agent name from the security context metadata.
func (i *Identity) TenantID() string { return i.ctx.Metadata.Agent }

// Email is unused in certificate-based auth.
func (i *Identity) Email() string { return "" }

// Scopes returns the permissions granted by the agent's certificate.
func (i *Identity) Scopes() []string { return i.ctx.Permissions }

// Roles is unused in certificate-based auth.
func (i *Identity) Roles() []string { return nil }

// HasScope checks whether the agent holds the given permission.
func (i *Identity) HasScope(s string) bool { return i.ctx.HasPermission(s) }

// HasRole is unused in certificate-based auth.
func (i *Identity) HasRole(_ string) bool { return false }

// Stats is unused in certificate-based auth.
func (i *Identity) Stats() map[string]int { return nil }

// Token returns the signed token for this identity.
func (i *Identity) Token() sctx.SignedToken { return i.token }

// Terminate returns rocco middleware that terminates mTLS. It extracts the
// client certificate, uses GenerateTrusted to issue a token (the TLS
// handshake already proved key possession), and injects the Identity into
// the request context. This middleware runs once, globally.
func Terminate(authority *Authority) func(http.Handler) http.Handler {
	admin := authority.Admin
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
				writeError(w, rocco.ErrUnauthorized.WithMessage("client certificate required"))
				return
			}

			cert := r.TLS.PeerCertificates[0]

			token, err := admin.GenerateTrusted(r.Context(), cert)
			if err != nil {
				writeError(w, rocco.ErrForbidden.WithMessage("certificate rejected"))
				return
			}

			fingerprint := sctx.GetFingerprint(cert)
			sctxCtx, ok := admin.GetContext(r.Context(), fingerprint)
			if !ok || sctxCtx.IsExpired() {
				writeError(w, rocco.ErrForbidden.WithMessage("security context unavailable"))
				return
			}

			identity := &Identity{ctx: sctxCtx, token: token}
			ctx := context.WithValue(r.Context(), identityKey{}, identity)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// IdentityFromContext extracts the Identity from the request context.
// Returns nil if Terminate middleware has not run.
func IdentityFromContext(ctx context.Context) *Identity {
	id, _ := ctx.Value(identityKey{}).(*Identity)
	return id
}

// Authenticator returns a rocco identity extractor that reads the Identity
// injected by the Terminate middleware.
func Authenticator() func(*http.Request) rocco.Identity {
	return func(r *http.Request) rocco.Identity {
		if id, ok := r.Context().Value(identityKey{}).(*Identity); ok {
			return id
		}
		return rocco.NoIdentity{}
	}
}

type identityKey struct{}
