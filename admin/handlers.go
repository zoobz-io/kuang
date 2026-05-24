package admin

import (
	"sync/atomic"

	"github.com/zoobz-io/kuang/internal/auth"
	"github.com/zoobz-io/kuang/internal/creds"
	"github.com/zoobz-io/rocco"
)

// ScopeAdmin is the scope required for all admin endpoints.
const ScopeAdmin = "admin"

func endpoints(store *creds.Store, setupEndpoint ...rocco.Endpoint) []rocco.Endpoint {
	eps := []rocco.Endpoint{
		listCredentials(store),
		getCredential(store),
		setCredential(store),
		deleteCredential(store),
	}
	eps = append(eps, setupEndpoint...)
	return eps
}

func setup(userStore *auth.Store) rocco.Endpoint {
	var done atomic.Bool
	return rocco.POST[SetupRequest, SetupResponse]("/setup", func(r *rocco.Request[SetupRequest]) (SetupResponse, error) {
		if done.Load() {
			return SetupResponse{}, rocco.ErrNotFound.WithMessage("setup already completed")
		}
		initialized, err := userStore.Initialized(r)
		if err != nil {
			return SetupResponse{}, rocco.ErrInternalServer.WithMessage("failed to check initialization state")
		}
		if initialized {
			done.Store(true)
			return SetupResponse{}, rocco.ErrNotFound.WithMessage("setup already completed")
		}
		if err := userStore.Initialize(r, r.Body.Username, r.Body.Password); err != nil {
			return SetupResponse{}, rocco.ErrInternalServer.WithMessage("failed to initialize")
		}
		done.Store(true)
		return SetupResponse{Username: r.Body.Username}, nil
	}).WithName("setup").WithSummary("First-time admin setup").WithTags("setup").WithErrors(rocco.ErrNotFound, rocco.ErrInternalServer)
}

func listCredentials(store *creds.Store) rocco.Endpoint {
	return rocco.GET[rocco.NoBody, CredentialKeyList]("/credentials/{agent}", func(r *rocco.Request[rocco.NoBody]) (CredentialKeyList, error) {
		agent := r.Params.Path["agent"]
		keys, err := store.List(r, agent)
		if err != nil {
			return CredentialKeyList{}, rocco.ErrInternalServer.WithMessage("failed to list credentials")
		}
		return CredentialKeyList{Agent: agent, Keys: keys}, nil
	}).WithName("listCredentials").WithSummary("List credential keys for an agent").WithTags("credentials").WithPathParams("agent").WithScopes(ScopeAdmin).WithErrors(rocco.ErrInternalServer)
}

func getCredential(store *creds.Store) rocco.Endpoint {
	return rocco.GET[rocco.NoBody, CredentialResponse]("/credentials/{agent}/{key}", func(r *rocco.Request[rocco.NoBody]) (CredentialResponse, error) {
		agent := r.Params.Path["agent"]
		key := r.Params.Path["key"]
		value, err := store.Resolve(r, agent, key)
		if err != nil {
			return CredentialResponse{}, rocco.ErrNotFound.WithMessage("credential not found")
		}
		return CredentialResponse{Agent: agent, Key: key, Value: value}, nil
	}).WithName("getCredential").WithSummary("Get a credential").WithTags("credentials").WithPathParams("agent", "key").WithScopes(ScopeAdmin).WithErrors(rocco.ErrNotFound)
}

func setCredential(store *creds.Store) rocco.Endpoint {
	return rocco.PUT[SetCredentialRequest, CredentialResponse]("/credentials/{agent}/{key}", func(r *rocco.Request[SetCredentialRequest]) (CredentialResponse, error) {
		agent := r.Params.Path["agent"]
		key := r.Params.Path["key"]
		if err := store.Set(r, agent, key, r.Body.Value); err != nil {
			return CredentialResponse{}, rocco.ErrInternalServer.WithMessage("failed to set credential")
		}
		return CredentialResponse{Agent: agent, Key: key, Value: r.Body.Value}, nil
	}).WithName("setCredential").WithSummary("Set a credential").WithTags("credentials").WithPathParams("agent", "key").WithScopes(ScopeAdmin).WithErrors(rocco.ErrInternalServer)
}

func deleteCredential(store *creds.Store) rocco.Endpoint {
	return rocco.DELETE[rocco.NoBody, rocco.NoBody]("/credentials/{agent}/{key}", func(r *rocco.Request[rocco.NoBody]) (rocco.NoBody, error) {
		agent := r.Params.Path["agent"]
		key := r.Params.Path["key"]
		if err := store.Delete(r, agent, key); err != nil {
			return rocco.NoBody{}, rocco.ErrInternalServer.WithMessage("failed to delete credential")
		}
		return rocco.NoBody{}, nil
	}).WithName("deleteCredential").WithSummary("Delete a credential").WithTags("credentials").WithPathParams("agent", "key").WithScopes(ScopeAdmin).WithErrors(rocco.ErrInternalServer)
}
