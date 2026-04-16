---
name: add-transformer
description: Create pure functions for mapping between models and wire types
---

# Add Transformer

You are creating transformers - pure functions that convert between internal models and API wire types. Transformers provide a clean separation between domain logic and API contracts.

## Surface Context

Transformers are surface-specific. Before proceeding:

1. **Determine the surface** — Is this for the public API or admin API?
2. **If unclear, ask** — "Which API surface: public (api/) or admin (admin/)?"
3. **Apply the correct path:**
   - Public API: `api/transformers/`
   - Admin API: `admin/transformers/`

## Technical Foundation

Transformers live in `{surface}/transformers/` as package-level functions. They have no side effects - no database calls, no external requests, just data mapping.

### Model to Response

Convert a domain model to an API response:

```go
package transformers

import (
    "github.com/yourorg/yourapp/models"
    "github.com/yourorg/yourapp/wire"
)

// UserToResponse transforms a User model to an API response.
func UserToResponse(u *models.User) wire.UserResponse {
    return wire.UserResponse{
        ID:        u.ID,
        Login:     u.Login,
        Email:     u.Email,
        Name:      u.Name,
        AvatarURL: u.AvatarURL,
    }
}
```

### Slice to List Response

Convert a slice of models to a list response:

```go
// RepositoriesToList transforms a slice of Repository models to an API list response.
func RepositoriesToList(repos []*models.Repository) wire.RepositoryListResponse {
    resp := wire.RepositoryListResponse{
        Repositories: make([]wire.RepositoryResponse, len(repos)),
    }
    for i, r := range repos {
        resp.Repositories[i] = RepositoryToResponse(r)
    }
    return resp
}
```

### Slice to Search Response

Convert search results with query context:

```go
// ChunksToSearchResponse transforms a slice of Chunk models to a search response.
func ChunksToSearchResponse(query string, chunks []*models.Chunk) wire.SearchResponse {
    resp := wire.SearchResponse{
        Query:   query,
        Results: make([]wire.ChunkResult, len(chunks)),
    }
    for i, c := range chunks {
        resp.Results[i] = ChunkToResult(c)
    }
    return resp
}
```

### Apply Request to Model

Apply a request's fields to an existing model (for updates):

```go
// ApplyUserUpdate applies a UserUpdateRequest to a User model.
func ApplyUserUpdate(req wire.UserUpdateRequest, u *models.User) {
    if req.Name != nil {
        u.Name = req.Name
    }
}
```

### Apply with Defaults

Apply request fields with default values for optional fields:

```go
// ApplyConfigRequest applies a ConfigRequest to a Config model.
func ApplyConfigRequest(req wire.ConfigRequest, c *models.Config) {
    c.Name = req.Name
    c.Enabled = req.Enabled

    // Apply optional field with default
    if req.MaxSize != nil {
        c.MaxSize = *req.MaxSize
    } else {
        c.MaxSize = models.DefaultMaxSize
    }
}
```

### Apply for Creation

Apply a request to create a new model:

```go
// ApplyRepositoryRegistration applies a RegisterRepositoryRequest to a Repository model.
func ApplyRepositoryRegistration(req wire.RegisterRepositoryRequest, r *models.Repository) {
    r.GitHubID = req.GitHubID
    r.Owner = req.Owner
    r.Name = req.Name
    r.FullName = req.FullName
    r.Description = req.Description
    r.DefaultBranch = req.DefaultBranch
    r.Private = req.Private
    r.HTMLURL = req.HTMLURL
}
```

### Filtering in Transformers

When transformation requires filtering:

```go
// DocumentsToSimilarResponse transforms documents, excluding a specific ID.
func DocumentsToSimilarResponse(docs []*models.Document, excludeID int64) wire.SimilarDocumentsResponse {
    resp := wire.SimilarDocumentsResponse{
        Results: make([]wire.DocumentResult, 0, len(docs)),
    }
    for _, d := range docs {
        if d.ID == excludeID {
            continue
        }
        resp.Results = append(resp.Results, DocumentToResult(d))
    }
    return resp
}
```

## Naming Conventions

| Pattern | Function Name | Example |
|---------|---------------|---------|
| Model → Response | `ModelToResponse` | `UserToResponse` |
| Model → Result | `ModelToResult` | `ChunkToResult` |
| Models → List | `ModelsToList` | `RepositoriesToList` |
| Models → Search | `ModelsToSearchResponse` | `ChunksToSearchResponse` |
| Apply Update | `ApplyModelUpdate` | `ApplyUserUpdate` |
| Apply Create | `ApplyModelRegistration` | `ApplyRepositoryRegistration` |
| Apply Config | `ApplyModelRequest` | `ApplyConfigRequest` |

## Guidelines

- **Pure functions only** - no side effects, no I/O, no database calls
- **One file per domain** - `users.go`, `repositories.go`, `search.go`
- **Accept pointers for models** - `*models.User` not `models.User`
- **Return values for responses** - `wire.UserResponse` not `*wire.UserResponse`
- **Mutate in place for Apply** - `ApplyUpdate(req, model)` modifies model
- **Handle nil gracefully** - check optional fields before dereferencing
- **Keep logic simple** - complex transformations may indicate a design issue

## Your Task

Understand what the user needs:

1. What domain entity is being transformed?
2. What direction: model → wire or wire → model?
3. Single item or collection?
4. For Apply: update existing or create new?
5. Are there optional fields needing defaults?
6. Any filtering or extra context (like query strings)?

## Before Writing Code

Produce a spec for approval:

```
## Transformer: [Domain]

**Surface:** [api / admin]

**File:** {surface}/transformers/[domain].go

**Functions:**

[FunctionName]([params]) [return type]
  Purpose: [What this function does]
  Direction: [model → wire / wire → model]

[NextFunction]...

**Special handling:**
- [Any defaults, filtering, or extra logic]
```

## After Approval

1. Create `{surface}/transformers/[domain].go` with imports
2. Implement each transformer function
3. Add doc comments explaining the transformation
4. If needed, create corresponding wire types (see `/add-wire`)
5. Use transformers in handlers (see `/add-handler`)

Replace `{surface}` with `api` or `admin` based on the target API surface.
