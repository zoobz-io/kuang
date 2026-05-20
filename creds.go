package kuang

import (
	"github.com/zoobz-io/check"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/kuang/contracts"
	"github.com/zoobz-io/kuang/internal/auth"
)

// ScopeCredsRead is the scope required to access the credentials endpoint.
const ScopeCredsRead = "creds-read"

// credentialResponse is the response for retrieving a credential.
type credentialResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Validate validates a credentialResponse.
func (c credentialResponse) Validate() error {
	return check.All(
		check.Str(c.Key, "key").Required().V(),
		check.Str(c.Value, "value").Required().V(),
	).Err()
}

// credsEndpoint returns the built-in GET /creds/{key} endpoint that lets
// agents retrieve their own credentials.
func credsEndpoint() rocco.Endpoint {
	return rocco.GET[rocco.NoBody, credentialResponse]("/creds/{key...}", func(r *rocco.Request[rocco.NoBody]) (credentialResponse, error) {
		creds, err := sum.Use[contracts.Credentials](r)
		if err != nil {
			return credentialResponse{}, rocco.ErrForbidden.WithMessage("credential store unavailable")
		}
		identity := auth.IdentityFromContext(r)
		if identity == nil {
			return credentialResponse{}, rocco.ErrUnauthorized.WithMessage("identity required")
		}
		key := r.Params.Path["key..."]
		value, err := creds.Resolve(r, identity.ID(), key)
		if err != nil {
			return credentialResponse{}, rocco.ErrForbidden.WithMessage("credential not found")
		}
		return credentialResponse{Key: key, Value: value}, nil
	}).WithName("getCredential").WithSummary("Get agent credential by key").WithTags("creds").WithPathParams("key...").WithScopes(ScopeCredsRead)
}
