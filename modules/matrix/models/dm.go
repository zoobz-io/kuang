package models

import "github.com/zoobz-io/check"

// DMSendRequest is the request body for sending a direct message.
type DMSendRequest struct {
	Message string `json:"message"`
}

// Validate validates a DMSendRequest.
func (r DMSendRequest) Validate() error {
	return check.All(
		check.Str(r.Message, "message").Required().V(),
	).Err()
}
