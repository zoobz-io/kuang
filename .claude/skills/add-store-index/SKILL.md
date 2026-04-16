---
name: add-store-index
description: Create a vector index store using grub.Index
---

# Add Store (Vector Index)

You are creating a vector index store - a typed wrapper around vector similarity search using `grub.Index[T]`. Use this for embeddings, semantic search, and recommendation systems.

## Shared Layer

Stores are **shared** across all API surfaces.

**Location:** `stores/`

**Registration:** `stores/stores.go`

## Technical Foundation

Index stores use:
- **grub** (`github.com/zoobzio/grub`) - `Index[T]` wrapper with atomization
- **vecna** (`github.com/zoobzio/vecna`) - Schema-validated filter builder

Review the GitHub repos if you need deeper understanding.

### Store Structure

```go
package stores

import (
    "context"

    "github.com/google/uuid"
    "github.com/zoobzio/grub"
    "github.com/zoobzio/vecna"
    "github.com/yourorg/yourapp/models"
)

// Embeddings provides vector similarity operations.
type Embeddings struct {
    index *grub.Index[models.EmbeddingMeta]
}

// NewEmbeddings creates a new embeddings store.
func NewEmbeddings(provider grub.VectorProvider) *Embeddings {
    return &Embeddings{index: grub.NewIndex[models.EmbeddingMeta](provider)}
}
```

### The Vector Type

grub.Index stores vectors with typed metadata:

```go
type Vector[T any] struct {
    ID       uuid.UUID // Unique identifier
    Vector   []float32 // The embedding
    Score    float32   // Similarity score (populated by Search)
    Metadata T         // Your typed metadata
}
```

### Operations

```go
// Upsert stores or updates a vector with metadata.
func (s *Embeddings) Upsert(ctx context.Context, id uuid.UUID, vector []float32, meta *models.EmbeddingMeta) error {
    return s.index.Upsert(ctx, id, vector, meta)
}

// Get retrieves a vector by ID.
func (s *Embeddings) Get(ctx context.Context, id uuid.UUID) (*grub.Vector[models.EmbeddingMeta], error) {
    return s.index.Get(ctx, id)
}

// Delete removes a vector.
func (s *Embeddings) Delete(ctx context.Context, id uuid.UUID) error {
    return s.index.Delete(ctx, id)
}

// Exists checks if a vector ID exists.
func (s *Embeddings) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
    return s.index.Exists(ctx, id)
}
```

### Similarity Search

```go
// Search finds the k nearest neighbors.
// filter is optional metadata filtering.
func (s *Embeddings) Search(ctx context.Context, vector []float32, k int, filter *models.EmbeddingMeta) ([]*grub.Vector[models.EmbeddingMeta], error) {
    return s.index.Search(ctx, vector, k, filter)
}
```

### Filtered Search with vecna

For complex metadata filters, use vecna:

```go
// QueryByUser finds similar vectors filtered by user.
func (s *Embeddings) QueryByUser(ctx context.Context, userID int64, vector []float32, k int) ([]*grub.Vector[models.EmbeddingMeta], error) {
    builder := vecna.New[models.EmbeddingMeta]()
    filter := builder.Where("user_id", vecna.Eq, userID)
    if err := filter.Err(); err != nil {
        return nil, err
    }
    return s.index.Query(ctx, vector, k, filter)
}

// QueryByUserAndType filters by multiple fields.
func (s *Embeddings) QueryByUserAndType(ctx context.Context, userID int64, contentType string, vector []float32, k int) ([]*grub.Vector[models.EmbeddingMeta], error) {
    builder := vecna.New[models.EmbeddingMeta]()
    filter := builder.And(
        builder.Where("user_id", vecna.Eq, userID),
        builder.Where("content_type", vecna.Eq, contentType),
    )
    if err := filter.Err(); err != nil {
        return nil, err
    }
    return s.index.Query(ctx, vector, k, filter)
}
```

### Batch Operations

```go
// UpsertBatch stores multiple vectors.
func (s *Embeddings) UpsertBatch(ctx context.Context, vectors []grub.Vector[models.EmbeddingMeta]) error {
    return s.index.UpsertBatch(ctx, vectors)
}

// DeleteBatch removes multiple vectors.
func (s *Embeddings) DeleteBatch(ctx context.Context, ids []uuid.UUID) error {
    return s.index.DeleteBatch(ctx, ids)
}
```

### Providers

Index stores are backed by a `grub.VectorProvider`. Available implementations:
- `grub/pinecone` - Pinecone
- `grub/qdrant` - Qdrant
- `grub/weaviate` - Weaviate
- `grub/milvus` - Milvus

## Your Task

Understand what the user is building:
1. What are the embeddings for? (documents, images, products)
2. What metadata needs to be stored with each vector?
3. What filters are needed for search? (by user, by type, by date)

If the metadata model doesn't exist yet, trigger `/add-model` first.

## Before Writing Code

Produce a spec for approval:

```
## Store: [Name]

**Metadata model:** [Model type for vector metadata]

**Vector dimensions:** [e.g., 1536 for OpenAI ada-002]

**Operations:**

Upsert(ctx, id, vector, metadata) error
Get(ctx, id) (*Vector[Model], error)
Delete(ctx, id) error

Search(ctx, vector, k, filter) ([]*Vector[Model], error)
  Basic similarity search

QueryBy[Filter](ctx, [filter params], vector, k) ([]*Vector[Model], error)
  Filtered search: [describe filter]

**Batch support:** [Yes/No]
```

## After Approval

1. If metadata model doesn't exist, trigger `/add-model` first
2. Create `stores/[name].go` with the store implementation
3. Update `stores/stores.go` to include the new store
