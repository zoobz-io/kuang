// Package contracts defines interfaces for kuang's core services.
package contracts

import (
	"context"
)

// Credentials manages credentials for authenticated agents.
// Modules use Resolve at request time to obtain the upstream credential
// for the calling agent (e.g. a GitHub token or Matrix access token).
// Agents use the /creds endpoints to manage the credentials they own.
type Credentials interface {
	// Resolve returns the credential value for the given agent and key.
	Resolve(ctx context.Context, agent string, key string) (string, error)
	// Set creates or updates the credential for the given agent and key.
	Set(ctx context.Context, agent string, key string, value string) error
	// Delete removes the credential for the given agent and key.
	Delete(ctx context.Context, agent string, key string) error
	// List returns the credential keys owned by the given agent.
	List(ctx context.Context, agent string) ([]string, error)
}
