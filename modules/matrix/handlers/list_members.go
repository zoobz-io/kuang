package handlers

import (
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobzio/kuang/core"
	"github.com/zoobzio/kuang/internal/httpc"
	"github.com/zoobzio/kuang/modules/matrix/contracts"
	"github.com/zoobzio/kuang/modules/matrix/models"
)

// ListMembers is the GET /matrix/rooms/{room}/members endpoint.
var ListMembers = core.GET[rocco.NoBody, models.MemberList]("/matrix/rooms/{room}/members", CredentialKey, func(r *rocco.Request[rocco.NoBody], auth httpc.RequestOption) (models.MemberList, error) {
	return sum.MustUse[contracts.API](r).Members(r, r.Params.Path["room"], auth)
}).WithName("listMembers").WithSummary("List room members").WithTags("matrix").WithPathParams("room").WithScopes(ScopeMembersRead)
