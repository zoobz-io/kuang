package models

import "github.com/zoobz-io/check"

// InviteRequest is the request body for inviting a user to a room.
type InviteRequest struct {
	UserID string `json:"user_id"`
}

// Validate validates an InviteRequest.
func (r InviteRequest) Validate() error {
	return check.All(
		check.Str(r.UserID, "user_id").Required().V(),
	).Err()
}

// Invite represents a pending room invite.
type Invite struct {
	RoomID string `json:"room_id"`
	Name   string `json:"name"`
	Sender string `json:"sender"`
}

// Validate validates an Invite.
func (i Invite) Validate() error { return nil }

// InviteList is the response for listing pending invites.
type InviteList struct {
	Invites []Invite `json:"invites"`
}

// Validate validates an InviteList.
func (i InviteList) Validate() error { return nil }
