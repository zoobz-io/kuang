// Package contracts defines interfaces for kuang's core services.
package contracts

import (
	"context"
)

// Credentials resolves credentials for authenticated agents.
// Modules use this at request time to obtain the upstream credential
// for the calling agent (e.g. a GitHub token or Matrix access token).
// Agents use this via the /creds endpoint to retrieve credentials
// they own.
type Credentials interface {
	// Resolve returns the credential value for the given agent and key.
	Resolve(ctx context.Context, agent string, key string) (string, error)
}
