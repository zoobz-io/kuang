package models

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
	NextBatch string       `json:"next_batch"`
	Events    []WatchEvent `json:"events"`
}

// Validate validates a WatchResponse.
func (w WatchResponse) Validate() error { return nil }
