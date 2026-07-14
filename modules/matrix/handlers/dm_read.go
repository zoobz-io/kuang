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

// DMRead is the GET /matrix/dm/{user}/messages endpoint.
var DMRead = core.GET[rocco.NoBody, models.MessageList]("/matrix/dm/{user}/messages", CredentialKey, func(r *rocco.Request[rocco.NoBody], auth httpc.RequestOption) (models.MessageList, error) {
	limit := 20
	if l := r.Params.Query["limit"]; l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}
	return sum.MustUse[contracts.API](r).DMRead(r, r.Params.Path["user"], limit, auth)
}).WithName("dmRead").WithSummary("Read DM history").WithTags("matrix").WithPathParams("user").WithQueryParams("limit").WithScopes(ScopeDMRead)
