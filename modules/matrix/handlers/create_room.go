package handlers

import (
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobzio/kuang/core"
	"github.com/zoobzio/kuang/internal/httpc"
	"github.com/zoobzio/kuang/modules/matrix/contracts"
	"github.com/zoobzio/kuang/modules/matrix/models"
)

// CreateRoom is the POST /matrix/rooms endpoint.
var CreateRoom = core.POST[models.CreateRoomRequest, models.Room]("/matrix/rooms", CredentialKey, func(r *rocco.Request[models.CreateRoomRequest], auth httpc.RequestOption) (models.Room, error) {
	return sum.MustUse[contracts.API](r).CreateRoom(r, r.Body.Name, r.Body.Topic, r.Body.Alias, auth)
}).WithName("createRoom").WithSummary("Create a room").WithTags("matrix").WithScopes(ScopeRoomsWrite)
