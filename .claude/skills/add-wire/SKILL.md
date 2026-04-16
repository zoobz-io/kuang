---
name: add-wire
description: Create API request and response types for HTTP handlers
---

# Add Wire

You are creating wire types - the structs that define the shape of data at the API boundary. Wire types are what clients send (requests) and receive (responses). They may differ from models (different fields, computed values, masked data).

## Surface Context

Wire types are surface-specific. Before proceeding:

1. **Determine the surface** — Is this for the public API or admin API?
2. **If unclear, ask** — "Which API surface: public (api/) or admin (admin/)?"
3. **Apply the correct path:**
   - Public API: `api/wire/`
   - Admin API: `admin/wire/`

**Key difference:** Public wire types use `OnSend()` boundaries for masking sensitive data. Admin wire types typically expose full data without masking.

## Technical Foundation

Wire types live in `{surface}/wire/` and implement rocco lifecycle interfaces for validation and boundary processing.

### Response Types

Response types are returned from handlers and may need masking before being sent to clients.

```go
package wire

import (
    "context"

    "github.com/zoobzio/sum"
)

// UserResponse is the API response for user data.
type UserResponse struct {
    ID        int64   `json:"id" description:"User ID" example:"12345"`
    Login     string  `json:"login" description:"GitHub username" example:"octocat"`
    Email     string  `json:"email" description:"Email address" example:"user@example.com" send.mask:"email"`
    Name      *string `json:"name,omitempty" description:"Display name" example:"Jane Doe" send.mask:"name"`
    AvatarURL *string `json:"avatar_url,omitempty" description:"Profile image URL"`
}

// OnSend applies boundary masking before the response is marshaled.
// Implements rocco.Sendable.
func (u *UserResponse) OnSend(ctx context.Context) error {
    b := sum.MustUse[*sum.Boundary[UserResponse]](ctx)
    masked, err := b.Send(ctx, *u)
    if err != nil {
        return err
    }
    *u = masked
    return nil
}

// Clone returns a deep copy. Required by cereal.Cloner.
func (u UserResponse) Clone() UserResponse {
    c := u
    if u.Name != nil {
        n := *u.Name
        c.Name = &n
    }
    if u.AvatarURL != nil {
        a := *u.AvatarURL
        c.AvatarURL = &a
    }
    return c
}
```

### Request Types

Request types are parsed from client input and need validation. They may also need boundary processing (e.g., password hashing).

```go
package wire

import "github.com/zoobzio/check"

// UserUpdateRequest is the request body for updating user profile.
type UserUpdateRequest struct {
    Name *string `json:"name,omitempty" description:"New display name" example:"Jane Doe"`
}

// Validate validates the request. Implements rocco.Validatable.
func (r *UserUpdateRequest) Validate() error {
    return check.All(
        check.OptStr(r.Name, "name").MaxLen(255).V(),
    ).Err()
}

// Clone returns a deep copy.
func (r UserUpdateRequest) Clone() UserUpdateRequest {
    c := r
    if r.Name != nil {
        n := *r.Name
        c.Name = &n
    }
    return c
}
```

### Request with Boundary Processing

When requests contain sensitive data that needs transformation (e.g., password hashing):

```go
package wire

import (
    "context"

    "github.com/zoobzio/check"
    "github.com/zoobzio/sum"
)

// RegisterRequest is the request body for user registration.
type RegisterRequest struct {
    Email    string `json:"email" description:"Email address" example:"user@example.com"`
    Password string `json:"password" description:"Password" receive.hash:"argon2"`
}

// Validate validates the request.
func (r *RegisterRequest) Validate() error {
    return check.All(
        check.Str(r.Email, "email").Required().Email().V(),
        check.Str(r.Password, "password").Required().MinLen(8).V(),
    ).Err()
}

// OnEntry applies boundary processing after validation.
// Implements rocco.Entryable.
func (r *RegisterRequest) OnEntry(ctx context.Context) error {
    b := sum.MustUse[*sum.Boundary[RegisterRequest]](ctx)
    processed, err := b.Receive(ctx, *r)
    if err != nil {
        return err
    }
    *r = processed
    return nil
}

// Clone returns a deep copy.
func (r RegisterRequest) Clone() RegisterRequest {
    return r // No reference fields
}
```

### List Response Types

For responses containing collections:

```go
// RepositoryListResponse is the API response for listing repositories.
type RepositoryListResponse struct {
    Repositories []RepositoryResponse `json:"repositories" description:"List of repositories"`
}

// Clone returns a deep copy.
func (r RepositoryListResponse) Clone() RepositoryListResponse {
    c := r
    if r.Repositories != nil {
        c.Repositories = make([]RepositoryResponse, len(r.Repositories))
        for i, repo := range r.Repositories {
            c.Repositories[i] = repo.Clone()
        }
    }
    return c
}
```

### Boundary Registration

Types that use boundary processing must be registered in `{surface}/wire/boundary.go`:

```go
package wire

import "github.com/zoobzio/sum"

// RegisterBoundaries creates and registers all wire boundaries.
// Call before sum.Freeze(k) in main.go.
func RegisterBoundaries(k sum.Key) error {
    // Response types with send.mask or send.redact
    if _, err := sum.NewBoundary[UserResponse](k); err != nil {
        return err
    }

    // Request types with receive.hash
    if _, err := sum.NewBoundary[RegisterRequest](k); err != nil {
        return err
    }

    return nil
}
```

### Struct Tags

| Tag | Purpose | Example |
|-----|---------|---------|
| `json:"field"` | JSON field name | `json:"user_id"` |
| `json:",omitempty"` | Omit if zero | `json:"name,omitempty"` |
| `description:"..."` | OpenAPI description | `description:"User ID"` |
| `example:"..."` | OpenAPI example | `example:"12345"` |
| `send.mask:"type"` | Mask on send | `send.mask:"email"` |
| `send.redact:"val"` | Redact on send | `send.redact:"***"` |
| `receive.hash:"algo"` | Hash on receive | `receive.hash:"argon2"` |

See `/add-boundary` for complete boundary tag documentation.

### Rocco Lifecycle Interfaces

| Interface | Method | Purpose |
|-----------|--------|---------|
| `Validatable` | `Validate() error` | Input/output validation |
| `Entryable` | `OnEntry(ctx) error` | Request boundary processing |
| `Sendable` | `OnSend(ctx) error` | Response boundary processing |

### Validation with check

Common validation patterns:

```go
// Required string
check.Str(r.Name, "name").Required().V()

// Optional string with max length
check.OptStr(r.Name, "name").MaxLen(255).V()

// Required positive integer
check.Int(r.ID, "id").Positive().V()

// Email validation
check.Str(r.Email, "email").Required().Email().V()

// URL validation
check.Str(r.URL, "url").Required().URL().V()

// One of allowed values
check.Str(r.Status, "status").Required().OneOf([]string{"active", "inactive"}).V()

// Combine multiple validations
check.All(
    check.Str(r.Name, "name").Required().MaxLen(255).V(),
    check.Str(r.Email, "email").Required().Email().V(),
).Err()
```

## Your Task

Understand what the user needs:

1. Is this a request or response type?
2. What fields does it have?
3. For requests: what validation rules apply?
4. For responses: are there sensitive fields needing masking?
5. Does the request need boundary processing (password hashing)?

## Before Writing Code

Produce a spec for approval:

```
## Wire Type: [Name]

**Kind:** [Request / Response]

**Purpose:** [What this type represents]

**Fields:**

[FieldName] ([type])
  json: [field name]
  description: [OpenAPI description]
  nullable: [yes/no]
  boundary: [send.mask:type / receive.hash:algo / none]

[NextField]...

**Validation:** [Rules using check, or "none" for responses]

**Boundary processing:** [OnSend / OnEntry / none]

**Needs boundary registration:** [yes/no]
```

## After Approval

1. Create `{surface}/wire/[domain].go` with the type(s)
2. Add `Clone()` method
3. For requests: add `Validate()` method using check
4. For responses with masking (public API): add `OnSend()` method
5. For requests with hashing: add `OnEntry()` method
6. If boundary processing used: register in `{surface}/wire/boundary.go`
7. Create transformer functions in `{surface}/transformers/`

Replace `{surface}` with `api` or `admin` based on the target API surface.

**Note:** Admin wire types typically don't need `OnSend()` boundaries — internal team sees full data.
