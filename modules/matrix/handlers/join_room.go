package handlers

import (
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobzio/kuang/core"
	"github.com/zoobzio/kuang/internal/httpc"
	"github.com/zoobzio/kuang/modules/matrix/contracts"
	"github.com/zoobzio/kuang/modules/matrix/models"
)

// JoinRoom is the POST /matrix/rooms/join endpoint.
var JoinRoom = core.POST[models.JoinRoomRequest, models.Room]("/matrix/rooms/join", CredentialKey, func(r *rocco.Request[models.JoinRoomRequest], auth httpc.RequestOption) (models.Room, error) {
	return sum.MustUse[contracts.API](r).Join(r, r.Body.RoomIDOrAlias, auth)
}).WithName("joinRoom").WithSummary("Join a room").WithTags("matrix").WithScopes(ScopeRoomsWrite)
