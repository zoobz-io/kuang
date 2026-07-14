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

// Watch is the GET /matrix/watch endpoint. It is the core endpoint for agent
// event loops, blocking until new messages arrive on any joined room or the
// timeout expires.
var Watch = core.GET[rocco.NoBody, models.WatchResponse]("/matrix/watch", CredentialKey, func(r *rocco.Request[rocco.NoBody], auth httpc.RequestOption) (models.WatchResponse, error) {
	since := r.Params.Query["since"]
	timeout := 30
	if t := r.Params.Query["timeout"]; t != "" {
		if n, err := strconv.Atoi(t); err == nil && n > 0 {
			timeout = n
		}
	}
	return sum.MustUse[contracts.API](r).Watch(r, since, timeout, auth)
}).WithName("watch").WithSummary("Watch for new messages (long-poll)").WithTags("matrix").WithQueryParams("since", "timeout").WithScopes(ScopeWatchRead)
