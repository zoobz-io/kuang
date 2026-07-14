// Package matrix provides a kuang module for Matrix messaging operations.
package matrix

import (
	"context"

	"github.com/zoobz-io/fig"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
	"github.com/zoobzio/kuang/core"
	"github.com/zoobzio/kuang/modules/matrix/contracts"
	"github.com/zoobzio/kuang/modules/matrix/handlers"
)

// Module returns a kuang module that provides Matrix messaging endpoints.
// The module registers the Matrix API service in the sum registry under the
// contracts.API contract and returns endpoints scoped by the matrix-*
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
