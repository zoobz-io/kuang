package github

import (
	"context"

	"github.com/zoobz-io/kuang"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

// Module returns a kuang module that provides GitHub API endpoints.
// The module creates an HTTP client for the GitHub API and registers
// endpoints scoped by the github-* permission family.
func Module(cfg Config) kuang.Module {
	return func(_ context.Context, _ sum.Key) ([]rocco.Endpoint, error) {
		if err := cfg.Validate(); err != nil {
			return nil, err
		}
		svc := newService(cfg)
		return endpoints(svc), nil
	}
}
