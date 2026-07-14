package handlers

import (
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobzio/kuang/core"
	"github.com/zoobzio/kuang/internal/httpc"
	"github.com/zoobzio/kuang/modules/github/contracts"
	"github.com/zoobzio/kuang/modules/github/models"
)

// CreateIssue is the POST /github/repos/{repo}/issues endpoint.
var CreateIssue = core.POST[models.CreateIssueRequest, models.Issue]("/github/repos/{repo}/issues", CredentialKey, func(r *rocco.Request[models.CreateIssueRequest], auth httpc.RequestOption) (models.Issue, error) {
	return sum.MustUse[contracts.API](r).CreateIssue(r, r.Params.Path["repo"], r.Body.Title, r.Body.Body, auth)
}).WithName("createIssue").WithSummary("Create issue").WithTags("github").WithPathParams("repo").WithScopes(ScopeIssuesWrite)
