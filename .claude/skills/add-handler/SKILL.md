---
name: add-handler
description: Create HTTP handlers using rocco for API endpoints
---

# Add Handler

You are creating HTTP handlers - the entry points for API requests. Handlers orchestrate calls to contracts and transform data between wire types and models.

## Surface Context

Handlers are surface-specific. Before proceeding:

1. **Determine the surface** — Is this for the public API or admin API?
2. **If unclear, ask** — "Which API surface: public (api/) or admin (admin/)?"
3. **Apply the correct path:**
   - Public API: `api/handlers/`
   - Admin API: `admin/handlers/`

Registration points vary by surface:
- Public: `api/handlers/handlers.go`, `api/handlers/errors.go`
- Admin: `admin/handlers/handlers.go`, `admin/handlers/errors.go`

## Technical Foundation

Handlers live in `{surface}/handlers/` as package-level variables using rocco's functional pattern. They are registered with the router via the `All()` function.

### Basic GET Handler

```go
package handlers

import (
    "github.com/zoobzio/rocco"
    "github.com/zoobzio/sum"
    "github.com/yourorg/yourapp/contracts"
    "github.com/yourorg/yourapp/transformers"
    "github.com/yourorg/yourapp/wire"
)

// GetMe returns the authenticated user's profile.
var GetMe = rocco.GET("/me", func(req *rocco.Request[rocco.NoBody]) (wire.UserResponse, error) {
    users := sum.MustUse[contracts.Users](req.Context)

    user, err := users.Get(req.Context, req.Identity.ID())
    if err != nil {
        return wire.UserResponse{}, err
    }

    return transformers.UserToResponse(user), nil
}).WithSummary("Get current user").
    WithDescription("Returns the authenticated user's profile.").
    WithTags("Users").
    WithAuthentication()
```

### POST Handler with Request Body

```go
// RegisterRepository registers a new repository.
var RegisterRepository = rocco.POST("/repositories", func(req *rocco.Request[wire.RegisterRepositoryRequest]) (wire.RepositoryResponse, error) {
    repos := sum.MustUse[contracts.Repositories](req.Context)

    userID, err := strconv.ParseInt(req.Identity.ID(), 10, 64)
    if err != nil {
        return wire.RepositoryResponse{}, err
    }

    repo := &models.Repository{UserID: userID}
    transformers.ApplyRepositoryRegistration(req.Body, repo)

    if err := repos.Set(req.Context, "", repo); err != nil {
        return wire.RepositoryResponse{}, err
    }

    return transformers.RepositoryToResponse(repo), nil
}).WithSummary("Register repository").
    WithDescription("Registers a new repository for processing.").
    WithTags("Repositories").
    WithAuthentication().
    WithSuccessStatus(201)
```

### PATCH Handler for Updates

```go
// UpdateMe updates the authenticated user's profile.
var UpdateMe = rocco.PATCH("/me", func(req *rocco.Request[wire.UserUpdateRequest]) (wire.UserResponse, error) {
    users := sum.MustUse[contracts.Users](req.Context)

    user, err := users.Get(req.Context, req.Identity.ID())
    if err != nil {
        return wire.UserResponse{}, err
    }

    transformers.ApplyUserUpdate(req.Body, user)

    if err := users.Set(req.Context, req.Identity.ID(), user); err != nil {
        return wire.UserResponse{}, err
    }

    return transformers.UserToResponse(user), nil
}).WithSummary("Update current user").
    WithDescription("Updates the authenticated user's profile.").
    WithTags("Users").
    WithAuthentication()
```

### Path Parameters

```go
// GetRepository returns a specific repository.
var GetRepository = rocco.GET("/repositories/{owner}/{repo}", func(req *rocco.Request[rocco.NoBody]) (wire.RepositoryResponse, error) {
    repos := sum.MustUse[contracts.Repositories](req.Context)

    owner := req.Params.Path["owner"]
    repoName := req.Params.Path["repo"]

    repo, err := repos.GetByOwnerAndName(req.Context, owner, repoName)
    if err != nil {
        return wire.RepositoryResponse{}, ErrRepositoryNotFound
    }

    return transformers.RepositoryToResponse(repo), nil
}).WithPathParams("owner", "repo").
    WithSummary("Get repository").
    WithDescription("Returns a specific repository.").
    WithTags("Repositories").
    WithErrors(ErrRepositoryNotFound).
    WithAuthentication()
```

### Query Parameters

```go
// SearchChunks performs semantic search.
var SearchChunks = rocco.GET("/search/{owner}/{repo}/{tag}", func(req *rocco.Request[rocco.NoBody]) (wire.SearchResponse, error) {
    chunks := sum.MustUse[contracts.Chunks](req.Context)

    // Required query parameter
    query := req.Params.Query["q"]
    if query == "" {
        return wire.SearchResponse{}, ErrMissingQuery
    }

    // Optional query parameter with default
    limit := 10
    if l := req.Params.Query["limit"]; l != "" {
        if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
            limit = parsed
        }
    }

    // Boolean query parameter
    exportedOnly := req.Params.Query["exported"] == "true"

    results, err := chunks.Search(req.Context, query, limit, exportedOnly)
    if err != nil {
        return wire.SearchResponse{}, err
    }

    return transformers.ChunksToSearchResponse(query, results), nil
}).WithPathParams("owner", "repo", "tag").
    WithQueryParams("q", "limit", "exported").
    WithSummary("Search chunks").
    WithDescription("Performs semantic search.").
    WithTags("Search").
    WithErrors(ErrMissingQuery).
    WithAuthentication()
```

### List Handler

```go
// ListRepositories returns all repositories for the user.
var ListRepositories = rocco.GET("/repositories", func(req *rocco.Request[rocco.NoBody]) (wire.RepositoryListResponse, error) {
    repos := sum.MustUse[contracts.Repositories](req.Context)

    userID, err := strconv.ParseInt(req.Identity.ID(), 10, 64)
    if err != nil {
        return wire.RepositoryListResponse{}, err
    }

    list, err := repos.ListByUserID(req.Context, userID)
    if err != nil {
        return wire.RepositoryListResponse{}, err
    }

    return transformers.RepositoriesToList(list), nil
}).WithSummary("List repositories").
    WithDescription("Returns all repositories for the authenticated user.").
    WithTags("Repositories").
    WithAuthentication()
```

### Error Definitions

Define domain-specific errors in `handlers/errors.go`:

```go
package handlers

import "github.com/zoobzio/rocco"

var (
    ErrRepositoryNotFound = rocco.ErrNotFound.WithMessage("repository not found")
    ErrVersionNotFound    = rocco.ErrNotFound.WithMessage("version not found")
    ErrMissingQuery       = rocco.ErrBadRequest.WithMessage("query parameter 'q' is required")
    ErrInvalidLimit       = rocco.ErrBadRequest.WithMessage("limit must be between 1 and 100")
)
```

Base errors from rocco:
- `rocco.ErrBadRequest` (400)
- `rocco.ErrUnauthorized` (401)
- `rocco.ErrForbidden` (403)
- `rocco.ErrNotFound` (404)
- `rocco.ErrConflict` (409)
- `rocco.ErrInternalServer` (500)

### Handler Registration

Register all handlers in `handlers/handlers.go`:

```go
package handlers

import "github.com/zoobzio/rocco"

// All returns all API handlers for registration.
func All() []rocco.Endpoint {
    return []rocco.Endpoint{
        // Users
        GetMe,
        UpdateMe,

        // Repositories
        ListRepositories,
        RegisterRepository,
        GetRepository,

        // Search
        SearchChunks,
    }
}
```

### Chainable Methods

| Method | Purpose | Example |
|--------|---------|---------|
| `.WithSummary()` | OpenAPI summary (short) | `WithSummary("Get user")` |
| `.WithDescription()` | OpenAPI description (detailed) | `WithDescription("Returns...")` |
| `.WithTags()` | OpenAPI tag grouping | `WithTags("Users")` |
| `.WithAuthentication()` | Require authentication | `WithAuthentication()` |
| `.WithPathParams()` | Document path variables | `WithPathParams("id", "name")` |
| `.WithQueryParams()` | Document query params | `WithQueryParams("q", "limit")` |
| `.WithErrors()` | Document expected errors | `WithErrors(ErrNotFound)` |
| `.WithSuccessStatus()` | Override default 200 | `WithSuccessStatus(201)` |

### HTTP Methods

| Method | Use Case | Example |
|--------|----------|---------|
| `rocco.GET` | Read operations | Fetch, list, search |
| `rocco.POST` | Create operations | Register, submit |
| `rocco.PATCH` | Partial updates | Update profile fields |
| `rocco.PUT` | Full replacement | Replace entire resource |
| `rocco.DELETE` | Remove operations | Delete resource |

### Request Types

| Type | Use Case |
|------|----------|
| `rocco.NoBody` | GET/DELETE with no body |
| `wire.SomeRequest` | POST/PATCH/PUT with body |

## Streaming Handlers (SSE)

For real-time server-to-client updates, use `rocco.NewStreamHandler`:

### Basic Stream Handler

```go
// StreamEvents streams real-time events to the client.
var StreamEvents = rocco.NewStreamHandler[rocco.NoBody, wire.EventUpdate](
    "event-stream",
    http.MethodGet,
    "/events/stream",
    func(req *rocco.Request[rocco.NoBody], stream rocco.Stream[wire.EventUpdate]) error {
        events := sum.MustUse[contracts.Events](req.Context)

        subscription := events.Subscribe(req.Context, req.Identity.ID())
        defer subscription.Close()

        // Keep-alive ticker
        keepAlive := time.NewTicker(30 * time.Second)
        defer keepAlive.Stop()

        for {
            select {
            case <-stream.Done():
                // Client disconnected
                return nil
            case <-keepAlive.C:
                stream.SendComment("keep-alive")
            case event := <-subscription.Events():
                if err := stream.Send(transformers.EventToUpdate(event)); err != nil {
                    return err
                }
            }
        }
    },
).WithSummary("Stream events").
    WithDescription("Real-time event stream via Server-Sent Events.").
    WithTags("Events").
    WithAuthentication()
```

### Stream with Path Parameters

```go
// StreamJobProgress streams progress updates for a specific job.
var StreamJobProgress = rocco.NewStreamHandler[rocco.NoBody, wire.ProgressUpdate](
    "job-progress-stream",
    http.MethodGet,
    "/jobs/{id}/progress",
    func(req *rocco.Request[rocco.NoBody], stream rocco.Stream[wire.ProgressUpdate]) error {
        jobs := sum.MustUse[contracts.Jobs](req.Context)

        jobID := req.Params.Path["id"]

        progress := jobs.WatchProgress(req.Context, jobID)
        for update := range progress {
            if err := stream.Send(wire.ProgressUpdate{
                JobID:    jobID,
                Progress: update.Percent,
                Status:   update.Status,
            }); err != nil {
                return err
            }
        }
        return nil
    },
).WithPathParams("id").
    WithSummary("Stream job progress").
    WithTags("Jobs").
    WithAuthentication()
```

### Stream with Request Body (POST)

```go
// StreamSearch streams search results as they're found.
var StreamSearch = rocco.NewStreamHandler[wire.StreamSearchRequest, wire.SearchResult](
    "search-stream",
    http.MethodPost,
    "/search/stream",
    func(req *rocco.Request[wire.StreamSearchRequest], stream rocco.Stream[wire.SearchResult]) error {
        search := sum.MustUse[contracts.Search](req.Context)

        results := search.StreamResults(req.Context, req.Body.Query, req.Body.Limit)
        for result := range results {
            if err := stream.Send(transformers.ResultToWire(result)); err != nil {
                return err
            }
        }
        return nil
    },
).WithSummary("Stream search results").
    WithTags("Search").
    WithAuthentication()
```

### Named Events

Use `SendEvent` to send typed events clients can filter:

```go
func(req *rocco.Request[rocco.NoBody], stream rocco.Stream[any]) error {
    // Client can listen for specific event types
    stream.SendEvent("progress", wire.ProgressUpdate{Percent: 50})
    stream.SendEvent("status", wire.StatusUpdate{Status: "processing"})
    stream.SendEvent("complete", wire.CompleteUpdate{Result: "done"})
    return nil
}
```

### Stream Interface

```go
type Stream[T any] interface {
    Send(data T) error                    // Send data-only event
    SendEvent(event string, data T) error // Send named event
    SendComment(comment string) error     // Send keep-alive comment
    Done() <-chan struct{}                // Client disconnect signal
}
```

### Stream Best Practices

- **Always check `stream.Done()`** - detect client disconnection
- **Use keep-alives** - send comments every 30s for long-lived streams
- **Clean up resources** - use `defer` to close subscriptions
- **Handle backpressure** - consider timeouts if events arrive faster than send

## Guidelines

- **Handlers are variables** - module-level `var Name = rocco.METHOD(...)`, not methods
- **Orchestration only** - handlers coordinate, don't contain business logic
- **Use contracts** - `sum.MustUse[contracts.T](req.Context)` for dependencies
- **Use transformers** - never manually map model ↔ wire in handlers
- **Return domain errors** - use errors from `errors.go`, not raw errors
- **Document everything** - summary, description, tags, errors for OpenAPI
- **Validate via wire types** - request bodies validate themselves (see `/add-wire`)

## Your Task

Understand what the user needs:

1. What HTTP method? (GET, POST, PATCH, PUT, DELETE)
2. What URL path? Any path parameters?
3. Does it need a request body?
4. What response type?
5. **Is this a streaming endpoint?** (real-time updates, progress, live feeds)
6. What contracts does it need?
7. What errors can it return?
8. Does it require authentication?

## Before Writing Code

Produce a spec for approval:

**For standard handlers:**

```
## Handler: [Name]

**Method:** [GET/POST/PATCH/PUT/DELETE]

**Path:** [/path/{param}]

**Request:** [rocco.NoBody or wire.SomeRequest]

**Response:** [wire.SomeResponse]

**Contracts:**
- [contracts.SomeContract] - [why needed]

**Path params:** [param1, param2] or "none"

**Query params:** [q, limit] or "none"

**Errors:**
- [ErrSomething] - [when returned]

**Authentication:** [required/optional/none]

**Success status:** [200/201/204]
```

**For streaming handlers:**

```
## Stream Handler: [Name]

**Method:** [GET/POST]

**Path:** [/path/{param}/stream]

**Request:** [rocco.NoBody or wire.SomeRequest]

**Event type:** [wire.SomeEvent]

**Named events:** [event1, event2] or "data only"

**Contracts:**
- [contracts.SomeContract] - [why needed]

**Path params:** [param1, param2] or "none"

**Query params:** [filter, limit] or "none"

**Keep-alive:** [yes/no]

**Authentication:** [required/optional/none]
```

## After Approval

**For standard handlers:**

1. Create handler in `{surface}/handlers/[domain].go`
2. Add any new errors to `{surface}/handlers/errors.go`
3. Add handler to the `All()` function in `{surface}/handlers/handlers.go`
4. Create wire types if needed (see `/add-wire`)
5. Create transformers if needed (see `/add-transformer`)

**For streaming handlers:**

1. Create stream handler in `{surface}/handlers/[domain].go` using `rocco.NewStreamHandler`
2. Add any new errors to `{surface}/handlers/errors.go`
3. Add handler to the `All()` function in `{surface}/handlers/handlers.go`
4. Create event wire types if needed (see `/add-wire`)
5. Create transformers for events if needed (see `/add-transformer`)
6. Implement proper cleanup with `defer` for subscriptions/resources

Replace `{surface}` with `api` or `admin` based on the target API surface.
