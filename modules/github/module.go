// Package github provides a kuang module for GitHub API operations.
package github

import (
	"context"

	"github.com/zoobz-io/fig"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
	"github.com/zoobzio/kuang/core"
	"github.com/zoobzio/kuang/modules/github/contracts"
	"github.com/zoobzio/kuang/modules/github/handlers"
)

// Module returns a kuang module that provides GitHub API endpoints.
// The module registers the GitHub API service in the sum registry under the
// contracts.API contract and returns endpoints scoped by the github-*
// permission family. Handlers resolve the service from the registry at
// request time.
func Module() core.Module {
	return func(_ context.Context, k sum.Key) ([]rocco.Endpoint, error) {
		var cfg Config
		if err := fig.Load(&cfg); err != nil {
			return nil, err
		}
		sum.Register[contracts.API](k, NewAPI(cfg))
		return handlers.Endpoints(), nil
	}
}
