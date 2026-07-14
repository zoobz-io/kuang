package handlers

import (
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobzio/kuang/core"
	"github.com/zoobzio/kuang/internal/httpc"
	"github.com/zoobzio/kuang/modules/matrix/contracts"
	"github.com/zoobzio/kuang/modules/matrix/models"
)

// ListRooms is the GET /matrix/rooms endpoint.
var ListRooms = core.GET[rocco.NoBody, models.RoomList]("/matrix/rooms", CredentialKey, func(r *rocco.Request[rocco.NoBody], auth httpc.RequestOption) (models.RoomList, error) {
	return sum.MustUse[contracts.API](r).JoinedRooms(r, auth)
}).WithName("listRooms").WithSummary("List joined rooms").WithTags("matrix").WithScopes(ScopeRoomsRead)
