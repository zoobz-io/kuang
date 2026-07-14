package models

import "github.com/zoobz-io/check"

// Identity is the response for the whoami endpoint.
type Identity struct {
	UserID string `json:"user_id"`
}

// Validate validates an Identity.
func (i Identity) Validate() error {
	return check.All(
		check.Str(i.UserID, "user_id").Required().V(),
	).Err()
}
