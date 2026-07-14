package handlers

import (
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobzio/kuang/core"
	"github.com/zoobzio/kuang/internal/httpc"
	"github.com/zoobzio/kuang/modules/github/contracts"
	"github.com/zoobzio/kuang/modules/github/models"
)

// GetRepo is the GET /github/repos/{name} endpoint.
var GetRepo = core.GET[rocco.NoBody, models.Repo]("/github/repos/{name}", CredentialKey, func(r *rocco.Request[rocco.NoBody], auth httpc.RequestOption) (models.Repo, error) {
	return sum.MustUse[contracts.API](r).GetRepo(r, r.Params.Path["name"], auth)
}).WithName("getRepo").WithSummary("Get repository").WithTags("github").WithPathParams("name").WithScopes(ScopeReposRead)
