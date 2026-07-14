package handlers

import (
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobzio/kuang/core"
	"github.com/zoobzio/kuang/internal/httpc"
	"github.com/zoobzio/kuang/modules/github/contracts"
	"github.com/zoobzio/kuang/modules/github/models"
)

// GetFile is the GET /github/repos/{repo}/content/{path...} endpoint.
var GetFile = core.GET[rocco.NoBody, models.FileContent]("/github/repos/{repo}/content/{path...}", CredentialKey, func(r *rocco.Request[rocco.NoBody], auth httpc.RequestOption) (models.FileContent, error) {
	return sum.MustUse[contracts.API](r).GetFile(r, r.Params.Path["repo"], r.Params.Path["path..."], r.Params.Query["ref"], auth)
}).WithName("getFile").WithSummary("Get file content").WithTags("github").WithPathParams("repo", "path...").WithQueryParams("ref").WithScopes(ScopeContentRead)
