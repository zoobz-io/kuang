package handlers

import (
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobzio/kuang/core"
	"github.com/zoobzio/kuang/internal/httpc"
	"github.com/zoobzio/kuang/modules/matrix/contracts"
	"github.com/zoobzio/kuang/modules/matrix/models"
)

// InviteUser is the POST /matrix/rooms/{room}/invite endpoint.
var InviteUser = core.POST[models.InviteRequest, rocco.NoBody]("/matrix/rooms/{room}/invite", CredentialKey, func(r *rocco.Request[models.InviteRequest], auth httpc.RequestOption) (rocco.NoBody, error) {
	return rocco.NoBody{}, sum.MustUse[contracts.API](r).Invite(r, r.Params.Path["room"], r.Body.UserID, auth)
}).WithName("inviteUser").WithSummary("Invite a user to a room").WithTags("matrix").WithPathParams("room").WithScopes(ScopeInvitesWrite)
