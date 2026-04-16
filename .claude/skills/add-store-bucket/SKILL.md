---
name: add-store-bucket
description: Create a blob store using grub.Bucket
---

# Add Store (Bucket)

You are creating a bucket store - a typed wrapper around blob/object storage using `grub.Bucket[T]`. Use this for files, serialized objects, or any data accessed by path.

## Shared Layer

Stores are **shared** across all API surfaces.

**Location:** `stores/`

**Registration:** `stores/stores.go`

## Technical Foundation

Bucket stores use:
- **grub** (`github.com/zoobzio/grub`) - `Bucket[T]` wrapper with atomization

Review the GitHub repo if you need deeper understanding.

### Store Structure

```go
package stores

import (
    "context"
    "fmt"

    "github.com/zoobzio/grub"
    "github.com/yourorg/yourapp/models"
)

// Blobs provides blob storage operations.
type Blobs struct {
    bucket *grub.Bucket[models.Blob]
}

// NewBlobs creates a new blob store.
func NewBlobs(provider grub.BucketProvider) *Blobs {
    return &Blobs{bucket: grub.NewBucket[models.Blob](provider)}
}
```

### Key Design

Bucket keys are paths. Design a consistent key structure:

```go
// Key builders for consistent path structure
func blobKey(userID int64, owner, repo, tag, path string) string {
    return fmt.Sprintf("%d/%s/%s/%s/%s", userID, owner, repo, tag, path)
}

func versionPrefix(userID int64, owner, repo, tag string) string {
    return fmt.Sprintf("%d/%s/%s/%s/", userID, owner, repo, tag)
}

func repoPrefix(userID int64, owner, repo string) string {
    return fmt.Sprintf("%d/%s/%s/", userID, owner, repo)
}
```

### Operations

```go
// Get retrieves a blob by path.
func (s *Blobs) GetByPath(ctx context.Context, userID int64, owner, repo, tag, path string) (*grub.Object[models.Blob], error) {
    return s.bucket.Get(ctx, blobKey(userID, owner, repo, tag, path))
}

// Put stores a blob.
func (s *Blobs) PutBlob(ctx context.Context, userID int64, blob *models.Blob) error {
    return s.bucket.Put(ctx, &grub.Object[models.Blob]{
        Key:         blobKey(userID, blob.Owner, blob.Repo, blob.Tag, blob.Path),
        ContentType: "application/json",
        Data:        *blob,
    })
}

// Delete removes a blob.
func (s *Blobs) DeleteByPath(ctx context.Context, userID int64, owner, repo, tag, path string) error {
    return s.bucket.Delete(ctx, blobKey(userID, owner, repo, tag, path))
}

// List returns objects matching a prefix.
func (s *Blobs) ListByVersion(ctx context.Context, userID int64, owner, repo, tag string, limit int) ([]grub.ObjectInfo, error) {
    return s.bucket.List(ctx, versionPrefix(userID, owner, repo, tag), limit)
}
```

### grub.Object[T]

Objects wrap data with metadata:

```go
type Object[T any] struct {
    Key         string            // Storage path
    ContentType string            // MIME type
    Size        int64             // Bytes (set by provider on Get)
    ETag        string            // Content hash (set by provider)
    Metadata    map[string]string // Custom metadata
    Data        T                 // The payload
}
```

### Providers

Bucket stores are backed by a `grub.BucketProvider`. Available implementations:
- `grub/minio` - MinIO/S3-compatible
- `grub/s3` - AWS S3
- `grub/gcs` - Google Cloud Storage
- `grub/azure` - Azure Blob Storage

## Your Task

Understand what the user is storing:
1. What kind of data? (files, serialized objects, binary blobs)
2. How will it be organized? (by user, by entity, by date)
3. What listing/prefix queries are needed?

If the model doesn't exist yet, trigger `/add-model` first.

## Before Writing Code

Produce a spec for approval:

```
## Store: [Name]

**Model:** [Model type for blob payload]

**Key structure:** [How keys are organized, e.g., "{user_id}/{owner}/{repo}/{path}"]

**Operations:**

GetByPath(ctx, [path params]) (*grub.Object[Model], error)
  Retrieves blob at path

PutBlob(ctx, [params], blob) error
  Stores blob at derived path

ListBy[Prefix](ctx, [prefix params], limit) ([]grub.ObjectInfo, error)
  Lists objects matching prefix
```

## After Approval

1. If model doesn't exist, trigger `/add-model` first
2. Create `stores/[name].go` with the store implementation
3. Add key builder functions for consistent paths
4. Update `stores/stores.go` to include the new store
