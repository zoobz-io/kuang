package admin

import "github.com/zoobz-io/check"

// SetCredentialRequest is the body for PUT /credentials/{agent}/{key}.
type SetCredentialRequest struct {
	Value string `json:"value"`
}

// Validate validates a SetCredentialRequest.
func (r SetCredentialRequest) Validate() error {
	return check.Check[SetCredentialRequest](
		check.Str(r.Value, "value").Required().V(),
	).Err()
}

// CredentialResponse is returned when retrieving a credential.
type CredentialResponse struct {
	Agent string `json:"agent"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Validate validates a CredentialResponse.
func (r CredentialResponse) Validate() error {
	return check.All(
		check.Str(r.Agent, "agent").Required().V(),
		check.Str(r.Key, "key").Required().V(),
	).Err()
}

// CredentialKeyList is returned when listing credential keys for an agent.
type CredentialKeyList struct {
	Agent string   `json:"agent"`
	Keys  []string `json:"keys"`
}

// Validate validates a CredentialKeyList.
func (r CredentialKeyList) Validate() error {
	return check.All(
		check.Str(r.Agent, "agent").Required().V(),
	).Err()
}
