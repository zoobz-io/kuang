// Package wire defines the HTTP request and response types for the
// credential API.
package wire

import (
	"github.com/zoobz-io/check"
)

// CredentialResponse is the response for retrieving a credential.
type CredentialResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Validate validates a CredentialResponse.
func (c CredentialResponse) Validate() error {
	return check.All(
		check.Str(c.Key, "key").Required().V(),
		check.Str(c.Value, "value").Required().V(),
	).Err()
}

// SetCredentialRequest is the request body for creating or updating a
// credential. The key is taken from the request path.
type SetCredentialRequest struct {
	Value string `json:"value"`
}

// Validate validates a SetCredentialRequest.
func (r SetCredentialRequest) Validate() error {
	return check.All(
		check.Str(r.Value, "value").Required().V(),
	).Err()
}

// CredentialKeysResponse is the response for listing an agent's credential
// keys. Values are never included in the listing.
type CredentialKeysResponse struct {
	Keys []string `json:"keys"`
}

// CredentialRef acknowledges a write or delete by echoing the affected key.
type CredentialRef struct {
	Key string `json:"key"`
}

// Validate validates a CredentialRef.
func (r CredentialRef) Validate() error {
	return check.All(
		check.Str(r.Key, "key").Required().V(),
	).Err()
}
