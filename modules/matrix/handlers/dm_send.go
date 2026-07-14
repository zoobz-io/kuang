package handlers

import (
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobzio/kuang/core"
	"github.com/zoobzio/kuang/internal/httpc"
	"github.com/zoobzio/kuang/modules/matrix/contracts"
	"github.com/zoobzio/kuang/modules/matrix/models"
)

// DMSend is the POST /matrix/dm/{user}/messages endpoint.
var DMSend = core.POST[models.DMSendRequest, models.SendMessageResponse]("/matrix/dm/{user}/messages", CredentialKey, func(r *rocco.Request[models.DMSendRequest], auth httpc.RequestOption) (models.SendMessageResponse, error) {
	return sum.MustUse[contracts.API](r).DMSend(r, r.Params.Path["user"], r.Body.Message, auth)
}).WithName("dmSend").WithSummary("Send a direct message").WithTags("matrix").WithPathParams("user").WithScopes(ScopeDMWrite)
