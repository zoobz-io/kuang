---
name: add-store-kv
description: Create a key-value store using grub.Store
---

# Add Store (Key-Value)

You are creating a key-value store - a typed wrapper around simple key→value storage using `grub.Store[T]`. Use this for caches, sessions, feature flags, or any data accessed by key with optional TTL.

## Shared Layer

Stores are **shared** across all API surfaces.

**Location:** `stores/`

**Registration:** `stores/stores.go`

## Technical Foundation

KV stores use:
- **grub** (`github.com/zoobzio/grub`) - `Store[T]` wrapper with atomization

Review the GitHub repo if you need deeper understanding.

### Store Structure

```go
package stores

import (
    "context"
    "time"

    "github.com/zoobzio/grub"
    "github.com/yourorg/yourapp/models"
)

// Sessions provides session storage with TTL support.
type Sessions struct {
    store *grub.Store[models.Session]
}

// NewSessions creates a new session store.
func NewSessions(provider grub.StoreProvider) *Sessions {
    return &Sessions{store: grub.NewStore[models.Session](provider)}
}
```

### Operations

```go
// Get retrieves a value by key.
func (s *Sessions) Get(ctx context.Context, key string) (*models.Session, error) {
    return s.store.Get(ctx, key)
}

// Set stores a value with TTL.
// TTL of 0 means no expiration.
func (s *Sessions) Set(ctx context.Context, key string, session *models.Session, ttl time.Duration) error {
    return s.store.Set(ctx, key, session, ttl)
}

// Delete removes a value.
func (s *Sessions) Delete(ctx context.Context, key string) error {
    return s.store.Delete(ctx, key)
}

// Exists checks if a key exists.
func (s *Sessions) Exists(ctx context.Context, key string) (bool, error) {
    return s.store.Exists(ctx, key)
}

// List returns keys matching a prefix.
func (s *Sessions) ListByUser(ctx context.Context, userID string, limit int) ([]string, error) {
    return s.store.List(ctx, "user:"+userID+":", limit)
}
```

### Batch Operations

For bulk reads/writes:

```go
// GetBatch retrieves multiple values.
// Missing keys are omitted from result.
func (s *Sessions) GetBatch(ctx context.Context, keys []string) (map[string]*models.Session, error) {
    return s.store.GetBatch(ctx, keys)
}

// SetBatch stores multiple values with TTL.
func (s *Sessions) SetBatch(ctx context.Context, sessions map[string]*models.Session, ttl time.Duration) error {
    return s.store.SetBatch(ctx, sessions, ttl)
}
```

### Providers

KV stores are backed by a `grub.StoreProvider`. Available implementations:
- `grub/redis` - Redis
- `grub/badger` - Badger (embedded)
- `grub/bolt` - BoltDB (embedded)

## Your Task

Understand what the user is storing:
1. What data? (sessions, cache entries, feature flags)
2. Does it need TTL? What expiration?
3. What key structure? (prefixes for listing)

If the model doesn't exist yet, trigger `/add-model` first.

## Before Writing Code

Produce a spec for approval:

```
## Store: [Name]

**Model:** [Model type for values]

**Key structure:** [How keys are organized, e.g., "session:{user_id}:{session_id}"]

**TTL:** [Default TTL, or "none"]

**Operations:**

Get(ctx, key) (*Model, error)
Set(ctx, key, value, ttl) error
Delete(ctx, key) error
ListBy[Prefix](ctx, [prefix params], limit) ([]string, error)

**Batch support:** [Yes/No]
```

## After Approval

1. If model doesn't exist, trigger `/add-model` first
2. Create `stores/[name].go` with the store implementation
3. Update `stores/stores.go` to include the new store
