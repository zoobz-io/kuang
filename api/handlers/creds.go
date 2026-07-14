// Package handlers defines the rocco endpoints for kuang's built-in
// per-agent credential API.
package handlers

import (
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobzio/kuang/api/contracts"
	"github.com/zoobzio/kuang/api/wire"
	"github.com/zoobzio/kuang/internal/auth"
)

// Scopes required to access the credentials endpoints. Read, write, and delete
// are separated so an agent can be granted, for example, read-only access.
const (
	ScopeCredsRead   = "creds-read"
	ScopeCredsWrite  = "creds-write"
	ScopeCredsDelete = "creds-delete"
)

// GetCredential is the GET /creds/{key} endpoint that lets agents retrieve
// one of their own credentials by key.
var GetCredential = rocco.GET[rocco.NoBody, wire.CredentialResponse]("/creds/{key...}", func(r *rocco.Request[rocco.NoBody]) (wire.CredentialResponse, error) {
	creds, err := sum.Use[contracts.Credentials](r)
	if err != nil {
		return wire.CredentialResponse{}, rocco.ErrForbidden.WithMessage("credential store unavailable")
	}
	identity := auth.IdentityFromContext(r)
	if identity == nil {
		return wire.CredentialResponse{}, rocco.ErrUnauthorized.WithMessage("identity required")
	}
	key := r.Params.Path["key..."]
	value, err := creds.Resolve(r, identity.ID(), key)
	if err != nil {
		return wire.CredentialResponse{}, rocco.ErrForbidden.WithMessage("credential not found")
	}
	return wire.CredentialResponse{Key: key, Value: value}, nil
}).WithName("getCredential").WithSummary("Get agent credential by key").WithTags("creds").WithPathParams("key...").WithScopes(ScopeCredsRead)

// ListCredentials is the GET /creds endpoint that lists the keys of the
// calling agent's credentials. Values are never returned by the listing.
var ListCredentials = rocco.GET[rocco.NoBody, wire.CredentialKeysResponse]("/creds", func(r *rocco.Request[rocco.NoBody]) (wire.CredentialKeysResponse, error) {
	creds, err := sum.Use[contracts.Credentials](r)
	if err != nil {
		return wire.CredentialKeysResponse{}, rocco.ErrForbidden.WithMessage("credential store unavailable")
	}
	identity := auth.IdentityFromContext(r)
	if identity == nil {
		return wire.CredentialKeysResponse{}, rocco.ErrUnauthorized.WithMessage("identity required")
	}
	keys, err := creds.List(r, identity.ID())
	if err != nil {
		return wire.CredentialKeysResponse{}, rocco.ErrInternalServer.WithMessage("failed to list credentials")
	}
	return wire.CredentialKeysResponse{Keys: keys}, nil
}).WithName("listCredentials").WithSummary("List agent credential keys").WithTags("creds").WithScopes(ScopeCredsRead)

// SetCredential is the PUT /creds/{key} endpoint that creates or updates one
// of the calling agent's credentials.
var SetCredential = rocco.PUT[wire.SetCredentialRequest, wire.CredentialRef]("/creds/{key...}", func(r *rocco.Request[wire.SetCredentialRequest]) (wire.CredentialRef, error) {
	creds, err := sum.Use[contracts.Credentials](r)
	if err != nil {
		return wire.CredentialRef{}, rocco.ErrForbidden.WithMessage("credential store unavailable")
	}
	identity := auth.IdentityFromContext(r)
	if identity == nil {
		return wire.CredentialRef{}, rocco.ErrUnauthorized.WithMessage("identity required")
	}
	key := r.Params.Path["key..."]
	if err := creds.Set(r, identity.ID(), key, r.Body.Value); err != nil {
		return wire.CredentialRef{}, rocco.ErrInternalServer.WithMessage("failed to store credential")
	}
	return wire.CredentialRef{Key: key}, nil
}).WithName("setCredential").WithSummary("Create or update an agent credential").WithTags("creds").WithPathParams("key...").WithScopes(ScopeCredsWrite)

// DeleteCredential is the DELETE /creds/{key} endpoint that removes one of the
// calling agent's credentials.
var DeleteCredential = rocco.DELETE[rocco.NoBody, wire.CredentialRef]("/creds/{key...}", func(r *rocco.Request[rocco.NoBody]) (wire.CredentialRef, error) {
	creds, err := sum.Use[contracts.Credentials](r)
	if err != nil {
		return wire.CredentialRef{}, rocco.ErrForbidden.WithMessage("credential store unavailable")
	}
	identity := auth.IdentityFromContext(r)
	if identity == nil {
		return wire.CredentialRef{}, rocco.ErrUnauthorized.WithMessage("identity required")
	}
	key := r.Params.Path["key..."]
	if err := creds.Delete(r, identity.ID(), key); err != nil {
		return wire.CredentialRef{}, rocco.ErrInternalServer.WithMessage("failed to delete credential")
	}
	return wire.CredentialRef{Key: key}, nil
}).WithName("deleteCredential").WithSummary("Delete an agent credential").WithTags("creds").WithPathParams("key...").WithScopes(ScopeCredsDelete)
