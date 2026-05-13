package matrix

import "github.com/zoobz-io/check"

// --- Rooms ---

// Room is a Matrix room identifier.
type Room struct {
	RoomID string `json:"room_id"`
}

// Validate validates a Room.
func (r Room) Validate() error {
	return check.All(
		check.Str(r.RoomID, "room_id").Required().V(),
	).Err()
}

// RoomInfo holds a room's name and topic.
type RoomInfo struct {
	RoomID string `json:"room_id"`
	Name   string `json:"name"`
	Topic  string `json:"topic"`
}

// Validate validates a RoomInfo.
func (r RoomInfo) Validate() error { return nil }

// RoomList is the response for listing joined rooms.
type RoomList struct {
	Rooms []RoomInfo `json:"rooms"`
}

// Validate validates a RoomList.
func (r RoomList) Validate() error { return nil }

// CreateRoomRequest is the request body for creating a room.
type CreateRoomRequest struct {
	Name  string `json:"name"`
	Topic string `json:"topic"`
	Alias string `json:"alias"`
}

// Validate validates a CreateRoomRequest.
func (r CreateRoomRequest) Validate() error {
	return check.Check[CreateRoomRequest](
		check.Str(r.Name, "name").Required().V(),
	).Err()
}

// JoinRoomRequest is the request body for joining a room.
type JoinRoomRequest struct {
	RoomIDOrAlias string `json:"room_id_or_alias"`
}

// Validate validates a JoinRoomRequest.
func (r JoinRoomRequest) Validate() error {
	return check.Check[JoinRoomRequest](
		check.Str(r.RoomIDOrAlias, "room_id_or_alias").Required().V(),
	).Err()
}

// --- Members ---

// Member represents a room member with their display name.
type Member struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
}

// Validate validates a Member.
func (m Member) Validate() error { return nil }

// MemberList is the response for listing room members.
type MemberList struct {
	Members []Member `json:"members"`
}

// Validate validates a MemberList.
func (m MemberList) Validate() error { return nil }

// --- Messages ---

// Message represents a Matrix timeline message.
type Message struct {
	Sender  string `json:"sender"`
	Body    string `json:"body"`
	EventID string `json:"event_id"`
	RoomID  string `json:"room_id,omitempty"`
}

// Validate validates a Message.
func (m Message) Validate() error { return nil }

// MessageList is the response for reading messages.
type MessageList struct {
	Messages []Message `json:"messages"`
}

// Validate validates a MessageList.
func (m MessageList) Validate() error { return nil }

// SendMessageRequest is the request body for sending a message.
type SendMessageRequest struct {
	Message string `json:"message"`
}

// Validate validates a SendMessageRequest.
func (r SendMessageRequest) Validate() error {
	return check.Check[SendMessageRequest](
		check.Str(r.Message, "message").Required().V(),
	).Err()
}

// SendMessageResponse is the response after sending a message.
type SendMessageResponse struct {
	EventID string `json:"event_id"`
}

// Validate validates a SendMessageResponse.
func (r SendMessageResponse) Validate() error {
	return check.All(
		check.Str(r.EventID, "event_id").Required().V(),
	).Err()
}

// --- Invites ---

// InviteRequest is the request body for inviting a user to a room.
type InviteRequest struct {
	UserID string `json:"user_id"`
}

// Validate validates an InviteRequest.
func (r InviteRequest) Validate() error {
	return check.Check[InviteRequest](
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

// --- DMs ---

// DMSendRequest is the request body for sending a direct message.
type DMSendRequest struct {
	Message string `json:"message"`
}

// Validate validates a DMSendRequest.
func (r DMSendRequest) Validate() error {
	return check.Check[DMSendRequest](
		check.Str(r.Message, "message").Required().V(),
	).Err()
}

// --- Watch ---

// WatchEvent represents a message or invite received during a watch.
type WatchEvent struct {
	Type     string `json:"type"`
	RoomID   string `json:"room_id"`
	RoomName string `json:"room_name,omitempty"`
	Sender   string `json:"sender"`
	Body     string `json:"body"`
	EventID  string `json:"event_id,omitempty"`
}

// Validate validates a WatchEvent.
func (w WatchEvent) Validate() error { return nil }

// WatchResponse is the response from the watch endpoint.
type WatchResponse struct {
	Events    []WatchEvent `json:"events"`
	NextBatch string       `json:"next_batch"`
}

// Validate validates a WatchResponse.
func (w WatchResponse) Validate() error { return nil }

// --- Identity ---

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
