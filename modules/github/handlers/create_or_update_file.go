package handlers

import (
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobzio/kuang/core"
	"github.com/zoobzio/kuang/internal/httpc"
	"github.com/zoobzio/kuang/modules/github/contracts"
	"github.com/zoobzio/kuang/modules/github/models"
)

// CreateOrUpdateFile is the PUT /github/repos/{repo}/content/{path...} endpoint.
var CreateOrUpdateFile = core.PUT[models.CreateOrUpdateFileRequest, models.FileContent]("/github/repos/{repo}/content/{path...}", CredentialKey, func(r *rocco.Request[models.CreateOrUpdateFileRequest], auth httpc.RequestOption) (models.FileContent, error) {
	return sum.MustUse[contracts.API](r).CreateOrUpdateFile(r, r.Params.Path["repo"], r.Params.Path["path..."], r.Body.Message, r.Body.Content, r.Body.SHA, auth)
}).WithName("createOrUpdateFile").WithSummary("Create or update file").WithTags("github").WithPathParams("repo", "path...").WithScopes(ScopeContentWrite)
