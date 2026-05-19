package secrets

import (
	"context"
	"fmt"

	"github.com/zoobz-io/fig"
	"github.com/zoobz-io/rocco"
)

// service wraps a fig.SecretProvider with scope-based key authorization.
type service struct {
	provider fig.SecretProvider
	// allowed maps each secret key to the set of scopes that grant access.
	allowed map[string][]string
}

func newService(cfg Config) *service {
	allowed := make(map[string][]string)
	for scope, keys := range cfg.Scopes {
		for _, key := range keys {
			allowed[key] = append(allowed[key], scope)
		}
	}
	return &service{provider: cfg.Provider, allowed: allowed}
}

// Get retrieves a secret by key, checking that the caller holds a scope
// granting access to that key.
func (s *service) Get(ctx context.Context, identity rocco.Identity, key string) (Secret, error) {
	if !s.authorize(identity, key) {
		return Secret{}, rocco.ErrForbidden.WithMessage(fmt.Sprintf("access denied for key %q", key))
	}
	value, err := s.provider.Get(ctx, key)
	if err != nil {
		return Secret{}, err
	}
	return Secret{Key: key, Value: value}, nil
}

// authorize checks whether the identity holds any scope that maps to the key.
func (s *service) authorize(identity rocco.Identity, key string) bool {
	scopes, ok := s.allowed[key]
	if !ok {
		return false
	}
	for _, scope := range scopes {
		if identity.HasScope(scope) {
			return true
		}
	}
	return false
}
