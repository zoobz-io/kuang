package handlers

import (
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobzio/kuang/core"
	"github.com/zoobzio/kuang/internal/httpc"
	"github.com/zoobzio/kuang/modules/github/contracts"
	"github.com/zoobzio/kuang/modules/github/models"
)

// ListPRs is the GET /github/repos/{repo}/pulls endpoint.
var ListPRs = core.GET[rocco.NoBody, models.PRList]("/github/repos/{repo}/pulls", CredentialKey, func(r *rocco.Request[rocco.NoBody], auth httpc.RequestOption) (models.PRList, error) {
	return sum.MustUse[contracts.API](r).ListPRs(r, r.Params.Path["repo"], auth)
}).WithName("listPRs").WithSummary("List pull requests").WithTags("github").WithPathParams("repo").WithScopes(ScopePullsRead)
