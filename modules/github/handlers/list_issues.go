package handlers

import (
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobzio/kuang/core"
	"github.com/zoobzio/kuang/internal/httpc"
	"github.com/zoobzio/kuang/modules/github/contracts"
	"github.com/zoobzio/kuang/modules/github/models"
)

// ListIssues is the GET /github/repos/{repo}/issues endpoint.
var ListIssues = core.GET[rocco.NoBody, models.IssueList]("/github/repos/{repo}/issues", CredentialKey, func(r *rocco.Request[rocco.NoBody], auth httpc.RequestOption) (models.IssueList, error) {
	return sum.MustUse[contracts.API](r).ListIssues(r, r.Params.Path["repo"], auth)
}).WithName("listIssues").WithSummary("List issues").WithTags("github").WithPathParams("repo").WithScopes(ScopeIssuesRead)
