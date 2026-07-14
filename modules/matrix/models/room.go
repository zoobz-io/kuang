// Package models defines the API request and response types for the Matrix module.
package models

import "github.com/zoobz-io/check"

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
	return check.All(
		check.Str(r.Name, "name").Required().V(),
	).Err()
}

// JoinRoomRequest is the request body for joining a room.
type JoinRoomRequest struct {
	RoomIDOrAlias string `json:"room_id_or_alias"`
}

// Validate validates a JoinRoomRequest.
func (r JoinRoomRequest) Validate() error {
	return check.All(
		check.Str(r.RoomIDOrAlias, "room_id_or_alias").Required().V(),
	).Err()
}
