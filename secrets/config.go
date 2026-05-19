// Package secrets provides a kuang module for agent-facing secret retrieval.
package secrets

import (
	"errors"
	"fmt"

	"github.com/zoobz-io/fig"
)

// Config holds the secret provider and scope-to-key mappings.
// Each key in Scopes is a scope string (e.g. "secrets-db") that maps to
// the secret keys an agent holding that scope may access.
type Config struct {
	Provider fig.SecretProvider
	Scopes   map[string][]string
}

// Validate checks that a provider and at least one scope mapping are set.
func (c Config) Validate() error {
	if c.Provider == nil {
		return errors.New("secrets: provider is required")
	}
	if len(c.Scopes) == 0 {
		return errors.New("secrets: at least one scope mapping is required")
	}
	for scope, keys := range c.Scopes {
		if len(keys) == 0 {
			return fmt.Errorf("secrets: scope %q has no keys", scope)
		}
	}
	return nil
}
