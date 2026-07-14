package handlers

import (
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobzio/kuang/core"
	"github.com/zoobzio/kuang/internal/httpc"
	"github.com/zoobzio/kuang/modules/matrix/contracts"
	"github.com/zoobzio/kuang/modules/matrix/models"
)

// WhoAmI is the GET /matrix/whoami endpoint.
var WhoAmI = core.GET[rocco.NoBody, models.Identity]("/matrix/whoami", CredentialKey, func(r *rocco.Request[rocco.NoBody], auth httpc.RequestOption) (models.Identity, error) {
	return sum.MustUse[contracts.API](r).WhoAmI(r, auth)
}).WithName("whoAmI").WithSummary("Get current identity").WithTags("matrix").WithScopes(ScopeIdentityRead)
