// Package contracts defines the interfaces for the Matrix module's services.
package contracts

import (
	"context"

	"github.com/zoobzio/kuang/internal/httpc"
	"github.com/zoobzio/kuang/modules/matrix/models"
)

// API describes the Matrix client-server API operations backing the module's
// endpoints. Handlers depend on this interface so the concrete service can be
// swapped or mocked in tests.
type API interface {
	// Identity.
	WhoAmI(ctx context.Context, opts ...httpc.RequestOption) (models.Identity, error)

	// Rooms.
	CreateRoom(ctx context.Context, name, topic, alias string, opts ...httpc.RequestOption) (models.Room, error)
	JoinedRooms(ctx context.Context, opts ...httpc.RequestOption) (models.RoomList, error)
	Join(ctx context.Context, roomIDOrAlias string, opts ...httpc.RequestOption) (models.Room, error)
	Leave(ctx context.Context, roomID string, opts ...httpc.RequestOption) error

	// Members.
	Members(ctx context.Context, roomID string, opts ...httpc.RequestOption) (models.MemberList, error)

	// Messages.
	Send(ctx context.Context, roomID, message string, opts ...httpc.RequestOption) (models.SendMessageResponse, error)
	ReadMessages(ctx context.Context, roomID string, limit int, from string, opts ...httpc.RequestOption) (models.MessageList, error)
	ReadMessagesSince(ctx context.Context, roomID, eventID string, limit int, from string, opts ...httpc.RequestOption) (models.MessageList, error)

	// Invites.
	Invite(ctx context.Context, roomID, userID string, opts ...httpc.RequestOption) error
	ListInvites(ctx context.Context, opts ...httpc.RequestOption) (models.InviteList, error)

	// DMs.
	DMSend(ctx context.Context, userID, message string, opts ...httpc.RequestOption) (models.SendMessageResponse, error)
	DMRead(ctx context.Context, userID string, limit int, opts ...httpc.RequestOption) (models.MessageList, error)

	// Watch.
	Watch(ctx context.Context, since string, timeout int, opts ...httpc.RequestOption) (models.WatchResponse, error)
}
