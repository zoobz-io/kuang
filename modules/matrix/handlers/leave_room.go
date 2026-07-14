package handlers

import (
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobzio/kuang/core"
	"github.com/zoobzio/kuang/internal/httpc"
	"github.com/zoobzio/kuang/modules/matrix/contracts"
)

// LeaveRoom is the POST /matrix/rooms/{room}/leave endpoint.
var LeaveRoom = core.POST[rocco.NoBody, rocco.NoBody]("/matrix/rooms/{room}/leave", CredentialKey, func(r *rocco.Request[rocco.NoBody], auth httpc.RequestOption) (rocco.NoBody, error) {
	return rocco.NoBody{}, sum.MustUse[contracts.API](r).Leave(r, r.Params.Path["room"], auth)
}).WithName("leaveRoom").WithSummary("Leave a room").WithTags("matrix").WithPathParams("room").WithScopes(ScopeRoomsWrite)
