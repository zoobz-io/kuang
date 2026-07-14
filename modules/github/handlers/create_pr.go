package handlers

import (
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobzio/kuang/core"
	"github.com/zoobzio/kuang/internal/httpc"
	"github.com/zoobzio/kuang/modules/github/contracts"
	"github.com/zoobzio/kuang/modules/github/models"
)

// CreatePR is the POST /github/repos/{repo}/pulls endpoint.
var CreatePR = core.POST[models.CreatePRRequest, models.PullRequest]("/github/repos/{repo}/pulls", CredentialKey, func(r *rocco.Request[models.CreatePRRequest], auth httpc.RequestOption) (models.PullRequest, error) {
	return sum.MustUse[contracts.API](r).CreatePR(r, r.Params.Path["repo"], r.Body.Title, r.Body.Body, r.Body.Head, r.Body.Base, auth)
}).WithName("createPR").WithSummary("Create pull request").WithTags("github").WithPathParams("repo").WithScopes(ScopePullsWrite)
