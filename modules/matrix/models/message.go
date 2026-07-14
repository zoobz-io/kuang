package models

import "github.com/zoobz-io/check"

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
	return check.All(
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
