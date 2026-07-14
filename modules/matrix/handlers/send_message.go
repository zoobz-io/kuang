package handlers

import (
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobzio/kuang/core"
	"github.com/zoobzio/kuang/internal/httpc"
	"github.com/zoobzio/kuang/modules/matrix/contracts"
	"github.com/zoobzio/kuang/modules/matrix/models"
)

// SendMessage is the POST /matrix/rooms/{room}/messages endpoint.
var SendMessage = core.POST[models.SendMessageRequest, models.SendMessageResponse]("/matrix/rooms/{room}/messages", CredentialKey, func(r *rocco.Request[models.SendMessageRequest], auth httpc.RequestOption) (models.SendMessageResponse, error) {
	return sum.MustUse[contracts.API](r).Send(r, r.Params.Path["room"], r.Body.Message, auth)
}).WithName("sendMessage").WithSummary("Send a message").WithTags("matrix").WithPathParams("room").WithScopes(ScopeMessagesWrite)
