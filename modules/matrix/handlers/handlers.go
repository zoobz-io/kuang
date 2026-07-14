// Package handlers defines the HTTP endpoints exposed by the Matrix module.
// Each endpoint resolves the API service from the sum registry at request time
// via sum.MustUse, and its bearer token via core's authenticated endpoint
// constructors.
package handlers

import "github.com/zoobz-io/rocco"

// Endpoints returns all HTTP endpoints exposed by the Matrix module.
func Endpoints() []rocco.Endpoint {
	return []rocco.Endpoint{
		WhoAmI,
		ListRooms, CreateRoom, JoinRoom, LeaveRoom,
		ListMembers,
		SendMessage, ReadMessages,
		InviteUser, ListInvites,
		DMSend, DMRead,
		Watch,
	}
}
