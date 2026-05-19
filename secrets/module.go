package secrets

import (
	"context"

	"github.com/zoobz-io/kuang"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

// Module returns a kuang module that provides scoped secret retrieval.
// The user defines scope-to-key mappings in Config.Scopes; an agent can
// only retrieve secrets whose keys are mapped to a scope the agent holds.
func Module(cfg Config) kuang.Module {
	return func(_ context.Context, _ sum.Key) ([]rocco.Endpoint, error) {
		if err := cfg.Validate(); err != nil {
			return nil, err
		}
		svc := newService(cfg)
		scopes := Scopes(cfg)
		return endpoints(svc, scopes), nil
	}
}
