package core

import (
	"context"

	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobzio/kuang/api/contracts"
	"github.com/zoobzio/kuang/internal/auth"
	"github.com/zoobzio/kuang/internal/httpc"
)

// ResolveBearer resolves the calling agent's bearer token for the given
// credential key from the registered credential store and returns it as an
// httpc request option. Modules use it to authenticate upstream API calls on
// behalf of the agent.
func ResolveBearer(ctx context.Context, credentialKey string) (httpc.RequestOption, error) {
	creds, err := sum.Use[contracts.Credentials](ctx)
	if err != nil {
		return nil, rocco.ErrForbidden.WithMessage("credential store unavailable")
	}
	identity := auth.IdentityFromContext(ctx)
	if identity == nil {
		return nil, rocco.ErrUnauthorized.WithMessage("identity required")
	}
	token, err := creds.Resolve(ctx, identity.ID(), credentialKey)
	if err != nil {
		return nil, rocco.ErrForbidden.WithMessage("credentials not configured")
	}
	return httpc.WithRequestBearerToken(token), nil
}

// GET defines an authenticated GET endpoint. The bearer token for
// credentialKey is resolved before fn runs and handed to fn as a request
// option to pass to upstream calls. It mirrors rocco.GET and returns the same
// *rocco.Handler, so the fluent builder (.WithName, .WithScopes, ...) chains
// as usual.
func GET[In, Out any](path, credentialKey string, fn func(*rocco.Request[In], httpc.RequestOption) (Out, error)) *rocco.Handler[In, Out] {
	return rocco.GET[In, Out](path, withBearer(credentialKey, fn))
}

// POST defines an authenticated POST endpoint. See GET.
func POST[In, Out any](path, credentialKey string, fn func(*rocco.Request[In], httpc.RequestOption) (Out, error)) *rocco.Handler[In, Out] {
	return rocco.POST[In, Out](path, withBearer(credentialKey, fn))
}

// PUT defines an authenticated PUT endpoint. See GET.
func PUT[In, Out any](path, credentialKey string, fn func(*rocco.Request[In], httpc.RequestOption) (Out, error)) *rocco.Handler[In, Out] {
	return rocco.PUT[In, Out](path, withBearer(credentialKey, fn))
}

// DELETE defines an authenticated DELETE endpoint. See GET.
func DELETE[In, Out any](path, credentialKey string, fn func(*rocco.Request[In], httpc.RequestOption) (Out, error)) *rocco.Handler[In, Out] {
	return rocco.DELETE[In, Out](path, withBearer(credentialKey, fn))
}

// withBearer adapts an auth-aware handler into a plain rocco handler by
// resolving the bearer token first and short-circuiting on failure.
func withBearer[In, Out any](credentialKey string, fn func(*rocco.Request[In], httpc.RequestOption) (Out, error)) func(*rocco.Request[In]) (Out, error) {
	return func(r *rocco.Request[In]) (Out, error) {
		var zero Out
		opt, err := ResolveBearer(r, credentialKey)
		if err != nil {
			return zero, err
		}
		return fn(r, opt)
	}
}
