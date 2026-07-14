package handlers

import (
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobzio/kuang/core"
	"github.com/zoobzio/kuang/internal/httpc"
	"github.com/zoobzio/kuang/modules/matrix/contracts"
	"github.com/zoobzio/kuang/modules/matrix/models"
)

// ListInvites is the GET /matrix/invites endpoint.
var ListInvites = core.GET[rocco.NoBody, models.InviteList]("/matrix/invites", CredentialKey, func(r *rocco.Request[rocco.NoBody], auth httpc.RequestOption) (models.InviteList, error) {
	return sum.MustUse[contracts.API](r).ListInvites(r, auth)
}).WithName("listInvites").WithSummary("List pending invites").WithTags("matrix").WithScopes(ScopeInvitesRead)
