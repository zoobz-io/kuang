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

// GetIssue is the GET /github/repos/{repo}/issues/{number} endpoint.
var GetIssue = core.GET[rocco.NoBody, models.Issue]("/github/repos/{repo}/issues/{number}", CredentialKey, func(r *rocco.Request[rocco.NoBody], auth httpc.RequestOption) (models.Issue, error) {
	num, err := strconv.Atoi(r.Params.Path["number"])
	if err != nil {
		return models.Issue{}, rocco.ErrBadRequest.WithMessage("invalid issue number")
	}
	return sum.MustUse[contracts.API](r).GetIssue(r, r.Params.Path["repo"], num, auth)
}).WithName("getIssue").WithSummary("Get issue").WithTags("github").WithPathParams("repo", "number").WithScopes(ScopeIssuesRead)
