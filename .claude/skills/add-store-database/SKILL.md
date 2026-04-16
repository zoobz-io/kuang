---
name: add-store-database
description: Create a SQL-backed store using sum.Database
---

# Add Store (Database)

You are creating a database store - a typed wrapper around SQL storage using `sum.Database[T]`. This is the most common store variant for structured, queryable data.

## Shared Layer

Stores are **shared** across all API surfaces. A single store can satisfy multiple contracts.

**Location:** `stores/`

**Registration:** `stores/stores.go`

Implement all query methods any surface might need. Different contracts expose different subsets.

## Technical Foundation

Database stores use these packages:
- **sum** (`github.com/zoobzio/sum`) - `Database[T]` wrapper (builds on grub)
- **soy** (`github.com/zoobzio/soy`) - Type-safe SQL query builder

Review the GitHub repos if you need deeper understanding.

### Store Structure

```go
package stores

import (
    "context"

    "github.com/jmoiron/sqlx"
    "github.com/zoobzio/astql"
    "github.com/zoobzio/sum"
    "github.com/yourorg/yourapp/models"
)

// Users provides database access for user records.
type Users struct {
    *sum.Database[models.User]
}

// NewUsers creates a new users store.
func NewUsers(db *sqlx.DB, renderer astql.Renderer) (*Users, error) {
    database, err := sum.NewDatabase[models.User](db, "users", renderer)
    if err != nil {
        return nil, err
    }
    return &Users{Database: database}, nil
}
```

### Built-in Operations

Embedding `*sum.Database[T]` provides these operations automatically:

```go
// Get by primary key
user, err := s.Get(ctx, key)

// Set (insert or update by primary key)
err := s.Set(ctx, key, user)

// Delete by primary key
err := s.Delete(ctx, key)

// Exists check
exists, err := s.Exists(ctx, key)
```

### Custom Queries with soy Builder

For queries beyond basic CRUD, use the builder shortcuts:

```go
// Select - returns single record
func (s *Users) GetByLogin(ctx context.Context, login string) (*models.User, error) {
    return s.Select().
        Where("login", "=", "login").
        Exec(ctx, map[string]any{"login": login})
}

// Query - returns multiple records
func (s *Documents) ListByVersion(ctx context.Context, userID int64, owner, repo, tag string) ([]*models.Document, error) {
    return s.Query().
        Where("user_id", "=", "user_id").
        Where("owner", "=", "owner").
        Where("repo_name", "=", "repo_name").
        Where("tag", "=", "tag").
        Exec(ctx, map[string]any{
            "user_id":   userID,
            "owner":     owner,
            "repo_name": repo,
            "tag":       tag,
        })
}

// Query with ordering and limit
func (s *Users) ListRecent(ctx context.Context, limit int) ([]*models.User, error) {
    return s.Query().
        OrderBy("created_at", "DESC").
        Limit(limit).
        Exec(ctx, nil)
}
```

**Builder shortcuts:** `s.Query()`, `s.Select()`, `s.Modify()`, `s.Remove()`, `s.Insert()`, `s.InsertFull()`

### Vector Similarity Queries

For models with `[]float32` vector fields:

```go
func (s *Documents) FindSimilar(ctx context.Context, userID int64, vector []float32, limit int) ([]*models.Document, error) {
    return s.Query().
        Where("user_id", "=", "user_id").
        OrderByExpr("vector", "<=>", "query_vec", "ASC").
        Limit(limit).
        Exec(ctx, map[string]any{
            "user_id":   userID,
            "query_vec": vector,
        })
}
```

### Transaction Support

Use `*Tx` method variants for transactional operations:

```go
func (s *Users) CreateWithProfile(ctx context.Context, tx *sqlx.Tx, user *models.User) error {
    return s.SetTx(ctx, tx, user.ID, user)
}
```

### Model Requirements

The model type T must have:
- A `db:"column"` tag on each persisted field
- A field with `constraints:"primarykey"` for the primary key
- See `/add-model` for full struct tag reference

### Lifecycle Hooks

If the model implements grub lifecycle interfaces, they're called automatically:
- `BeforeSave(ctx) error` - before insert/update
- `AfterLoad(ctx) error` - after select
- `AfterSave(ctx) error` - after insert/update

## Your Task

Understand what the user is storing:
1. What entity/model does this store manage?
2. What queries does the application need beyond basic CRUD?
3. Does this store need transaction support?

If the model doesn't exist yet, trigger `/add-model` first.

## Before Writing Code

Produce a spec for approval:

```
## Store: [Name]

**Model:** [Model type this store manages]

**Table:** [Database table name]

**Operations:**

Get(ctx, key) (*Model, error)
  Built-in primary key lookup

[CustomMethod](ctx, [params]) ([returns], error)
  Purpose: [What this query does]
  Query: [Select/Query, which fields, conditions, ordering]

**Transactions:** [Yes/No - does any operation need tx support?]
```

## After Approval

1. If model doesn't exist, trigger `/add-model` first
2. Create `stores/[name].go` with the store implementation
3. Update `stores/stores.go` to include the new store in the aggregate
4. Register the store against surface-specific contracts in each binary's main.go
