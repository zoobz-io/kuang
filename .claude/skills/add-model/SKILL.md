---
name: add-model
description: Create a domain model type in the models/ package
---

# Add Model

You are creating a domain model - a struct that represents an entity in the system. Models are pure data types that flow through stores, contracts, and handlers.

## Technical Foundation

Models live in `models/` as plain structs. The struct tags you need depend on what kind of store will hold this model.

### For Database Stores (SQL)

Models stored in `sum.Database[T]` need soy-compatible struct tags:

```go
package models

import "time"

// User represents an authenticated user.
type User struct {
    ID        int64     `json:"id" db:"id" constraints:"primarykey" description:"User ID" example:"12345"`
    Email     string    `json:"email" db:"email" constraints:"notnull,unique" description:"User email" example:"user@example.com"`
    Name      *string   `json:"name,omitempty" db:"name" description:"Display name" example:"Jane Doe"`
    CreatedAt time.Time `json:"created_at" db:"created_at" default:"now()" description:"Account creation time"`
}
```

**Database struct tags** (used by `github.com/zoobzio/soy`):

| Tag | Purpose | Example |
|-----|---------|---------|
| `db:"column"` | Column name (required) | `db:"user_id"` |
| `db:"-"` | Exclude from database | `db:"-"` |
| `constraints:"..."` | SQL constraints | `constraints:"primarykey,notnull,unique"` |
| `references:"..."` | Foreign key | `references:"users(id)"` |
| `default:"..."` | Default value | `default:"now()"` or `default:"0"` |
| `type:"..."` | Explicit SQL type | `type:"text"` (usually inferred) |
| `index:"..."` | Index name | `index:"idx_users_email"` |
| `check:"..."` | Check constraint | `check:"age > 0"` |
| `description:"..."` | Field documentation (DBML) | `description:"GitHub user ID"` |
| `example:"..."` | Example value (DBML) | `example:"12345678"` |

### For Other Stores (Bucket, KV, Index)

Models stored in `grub.Bucket[T]`, `grub.Store[T]`, or `grub.Index[T]` only need JSON tags:

```go
// Session is stored in a KV store.
type Session struct {
    UserID    int64     `json:"user_id"`
    Token     string    `json:"token"`
    ExpiresAt time.Time `json:"expires_at"`
}

// EmbeddingMeta is metadata for vectors in an Index store.
type EmbeddingMeta struct {
    DocumentID int64  `json:"document_id"`
    ChunkID    int64  `json:"chunk_id"`
    UserID     int64  `json:"user_id"`
}
```

### Common Patterns

**JSON tags** (all store types):

```go
Email string `json:"email"`           // normal field
Token string `json:"-"`               // omit from JSON (internal only)
Name  string `json:"name,omitempty"`  // omit if zero value
```

**Nullable fields** - use pointers:

```go
Name      *string    `json:"name,omitempty" db:"name"`
DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
```

**Vector fields** - for embeddings:

```go
Vector []float32 `json:"-" db:"vector"` // typically omitted from JSON
```

**Field-level encryption** - for sensitive data:

```go
AccessToken string `json:"-" db:"access_token" store.encrypt:"aes" load.decrypt:"aes"`
```

### Lifecycle Hooks

Models can implement grub lifecycle interfaces (called automatically by stores):

```go
import "context"

func (u *User) BeforeSave(ctx context.Context) error {
    // Called before insert/update
    return nil
}

func (u *User) AfterLoad(ctx context.Context) error {
    // Called after select/get
    return nil
}
```

Available hooks:
- `BeforeSave(ctx context.Context) error`
- `AfterSave(ctx context.Context) error`
- `AfterLoad(ctx context.Context) error`
- `BeforeDelete(ctx context.Context) error`
- `AfterDelete(ctx context.Context) error`

### Lifecycle Hooks with Boundary Processing

When models have encrypted fields, use lifecycle hooks with boundaries:

```go
package models

import (
    "context"

    "github.com/zoobzio/sum"
)

type OAuthToken struct {
    ID          int64  `json:"id" db:"id" constraints:"primarykey"`
    UserID      int64  `json:"user_id" db:"user_id" constraints:"notnull"`
    AccessToken string `json:"-" db:"access_token" store.encrypt:"aes" load.decrypt:"aes"`
}

// BeforeSave encrypts sensitive fields before database write.
func (t *OAuthToken) BeforeSave(ctx context.Context) error {
    b := sum.MustUse[*sum.Boundary[OAuthToken]](ctx)
    stored, err := b.Store(ctx, *t)
    if err != nil {
        return err
    }
    *t = stored
    return nil
}

// AfterLoad decrypts sensitive fields after database read.
func (t *OAuthToken) AfterLoad(ctx context.Context) error {
    b := sum.MustUse[*sum.Boundary[OAuthToken]](ctx)
    loaded, err := b.Load(ctx, *t)
    if err != nil {
        return err
    }
    *t = loaded
    return nil
}

func (t OAuthToken) Clone() OAuthToken { return t }
```

Models with boundary processing must be registered. See `/add-boundary` for complete documentation.

### Clone Method

Models should have a `Clone()` method for deep copying:

```go
func (u User) Clone() User {
    c := u
    if u.Name != nil {
        n := *u.Name
        c.Name = &n
    }
    if u.Vector != nil {
        c.Vector = make([]float32, len(u.Vector))
        copy(c.Vector, u.Vector)
    }
    return c
}
```

### Validate Method

Models should have a `Validate()` method using `github.com/zoobzio/check`:

```go
func (u User) Validate() error {
    return check.All(
        check.Str(u.Login, "login").Required().MaxLen(255).V(),
        check.Str(u.Email, "email").Required().V(),
        check.Int(u.ID, "id").Positive().V(),
    ).Err()
}
```

### Atomization (Optional)

For performance-critical models, implement the `atom.Atomizable` and `atom.Deatomizable` interfaces to avoid reflection:

```go
import "github.com/zoobzio/atom"

func (u *User) Atomize(a *atom.Atom) {
    a.Ints["id"] = u.ID
    a.Strings["email"] = u.Email
    // ... map all fields to atom
}

func (u *User) Deatomize(a *atom.Atom) error {
    u.ID = a.Ints["id"]
    u.Email = a.Strings["email"]
    // ... restore all fields from atom
    return nil
}
```

This is optional but recommended for high-throughput models.

## Your Task

Understand what the user is modeling:
1. What entity does this represent?
2. What kind of store will hold it? (Database, Bucket, KV, Index)
3. What fields does it have?
4. Which fields are nullable?
5. Are there sensitive fields needing encryption?
6. Does it need lifecycle hooks?
7. What validation rules apply?
8. Is this a high-throughput model needing atomization?

## Before Writing Code

Produce a spec for approval:

```
## Model: [Name]

**Purpose:** [What this entity represents]

**Store type:** [Database/Bucket/KV/Index]

**Fields:**

[FieldName] ([type])
  json: [field name or "-"]
  db: [column name] (if Database)
  constraints: [primarykey/notnull/unique] (if Database)
  nullable: [yes/no]

[NextField]...

**Lifecycle hooks:** [BeforeSave/AfterLoad/etc. or "none"]

**Encryption:** [Fields needing encryption or "none"]

**Validation:** [Validation rules or "none"]

**Atomization:** [Yes (for high-throughput models) or "no"]
```

## After Approval

1. Create `models/[name].go` with the struct and doc comment
2. Add `Clone()` method
3. Add `Validate()` method if validation rules apply
4. Add lifecycle hooks if needed (BeforeSave/AfterLoad for boundary processing)
5. If encrypted fields: register boundary in `models/boundary.go` (see `/add-boundary`)
6. Add `Atomize()`/`Deatomize()` methods if high-throughput
7. Proceed to `/add-store` to create the store that holds this model
8. If Database store, a migration will be needed (see `/add-migration`)
