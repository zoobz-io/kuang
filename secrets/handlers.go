package secrets

import (
	"sort"

	"github.com/zoobz-io/kuang/internal/auth"
	"github.com/zoobz-io/rocco"
)

// Scopes returns all scope strings defined by this module's configuration.
func Scopes(cfg Config) []string {
	seen := make(map[string]bool)
	for scope := range cfg.Scopes {
		seen[scope] = true
	}
	out := make([]string, 0, len(seen))
	for s := range seen {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

func endpoints(svc *service, scopes []string) []rocco.Endpoint {
	return []rocco.Endpoint{
		getSecret(svc, scopes),
	}
}

func getSecret(svc *service, scopes []string) rocco.Endpoint {
	return rocco.GET[rocco.NoBody, Secret]("/secrets/{key...}", func(r *rocco.Request[rocco.NoBody]) (Secret, error) {
		identity := auth.IdentityFromContext(r)
		if identity == nil {
			return Secret{}, rocco.ErrUnauthorized.WithMessage("identity required")
		}
		return svc.Get(r, identity, r.Params.Path["key..."])
	}).WithName("getSecret").WithSummary("Get secret by key").WithTags("secrets").WithPathParams("key...").WithScopes(scopes...)
}
