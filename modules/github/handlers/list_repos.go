package handlers

import (
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobzio/kuang/core"
	"github.com/zoobzio/kuang/internal/httpc"
	"github.com/zoobzio/kuang/modules/github/contracts"
	"github.com/zoobzio/kuang/modules/github/models"
)

// ListRepos is the GET /github/repos endpoint.
var ListRepos = core.GET[rocco.NoBody, models.RepoList]("/github/repos", CredentialKey, func(r *rocco.Request[rocco.NoBody], auth httpc.RequestOption) (models.RepoList, error) {
	return sum.MustUse[contracts.API](r).ListRepos(r, auth)
}).WithName("listRepos").WithSummary("List repositories").WithTags("github").WithScopes(ScopeReposRead)
