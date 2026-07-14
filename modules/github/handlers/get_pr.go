package handlers

import (
	"strconv"

	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobzio/kuang/core"
	"github.com/zoobzio/kuang/internal/httpc"
	"github.com/zoobzio/kuang/modules/github/contracts"
	"github.com/zoobzio/kuang/modules/github/models"
)

// GetPR is the GET /github/repos/{repo}/pulls/{number} endpoint.
var GetPR = core.GET[rocco.NoBody, models.PullRequest]("/github/repos/{repo}/pulls/{number}", CredentialKey, func(r *rocco.Request[rocco.NoBody], auth httpc.RequestOption) (models.PullRequest, error) {
	num, err := strconv.Atoi(r.Params.Path["number"])
	if err != nil {
		return models.PullRequest{}, rocco.ErrBadRequest.WithMessage("invalid PR number")
	}
	return sum.MustUse[contracts.API](r).GetPR(r, r.Params.Path["repo"], num, auth)
}).WithName("getPR").WithSummary("Get pull request").WithTags("github").WithPathParams("repo", "number").WithScopes(ScopePullsRead)
