package handlers

import (
	"strconv"

	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobzio/kuang/core"
	"github.com/zoobzio/kuang/internal/httpc"
	"github.com/zoobzio/kuang/modules/matrix/contracts"
	"github.com/zoobzio/kuang/modules/matrix/models"
)

// ReadMessages is the GET /matrix/rooms/{room}/messages endpoint.
var ReadMessages = core.GET[rocco.NoBody, models.MessageList]("/matrix/rooms/{room}/messages", CredentialKey, func(r *rocco.Request[rocco.NoBody], auth httpc.RequestOption) (models.MessageList, error) {
	limit := 20
	if l := r.Params.Query["limit"]; l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}
	since := r.Params.Query["since"]
	from := r.Params.Query["from"]
	svc := sum.MustUse[contracts.API](r)
	if since != "" {
		return svc.ReadMessagesSince(r, r.Params.Path["room"], since, limit, from, auth)
	}
	return svc.ReadMessages(r, r.Params.Path["room"], limit, from, auth)
}).WithName("readMessages").WithSummary("Read messages from a room").WithTags("matrix").WithPathParams("room").WithQueryParams("limit", "since", "from").WithScopes(ScopeMessagesRead)
