package models

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
