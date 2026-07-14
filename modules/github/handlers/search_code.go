package handlers

import (
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobzio/kuang/core"
	"github.com/zoobzio/kuang/internal/httpc"
	"github.com/zoobzio/kuang/modules/github/contracts"
	"github.com/zoobzio/kuang/modules/github/models"
)

// SearchCode is the GET /github/search/code endpoint.
var SearchCode = core.GET[rocco.NoBody, models.CodeSearchResult]("/github/search/code", CredentialKey, func(r *rocco.Request[rocco.NoBody], auth httpc.RequestOption) (models.CodeSearchResult, error) {
	return sum.MustUse[contracts.API](r).SearchCode(r, r.Params.Query["query"], auth)
}).WithName("searchCode").WithSummary("Search code").WithTags("github").WithQueryParams("query").WithScopes(ScopeSearchRead)
