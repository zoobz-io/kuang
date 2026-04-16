---
name: add-store
description: Create a data store backed by grub
---

# Add Store

You are creating a store - a typed wrapper around a data source. All stores in this application use `github.com/zoobzio/grub` which provides provider-agnostic storage with atomization support.

## Shared Layer

Stores are **shared** across all API surfaces. A single store implementation can satisfy multiple contracts — one for the public API, one for the admin API.

**Location:** `stores/`

**Registration:** `stores/stores.go` (aggregate factory)

The same `stores.Users` implementation might satisfy:
- `api/contracts.Users` (Get, Set, GetByLogin)
- `admin/contracts.Users` (Get, Set, Delete, List, Count)

Implement all methods any surface might need. Each contract exposes only what that surface requires.

## Storage Variants

Grub supports four storage variants. Understand what the user is storing to pick the right one:

| Variant | URI | Use Case | Grub Type |
|---------|-----|----------|-----------|
| **Database** | `db://` | Structured records with SQL queries | `grub.Database[T]` |
| **Bucket** | `bcs://` | Blobs/files with metadata | `grub.Bucket[T]` |
| **Key-Value** | `kv://` | Simple key→value with optional TTL | `grub.Store[T]` |
| **Index** | `idx://` | Vectors for similarity search | `grub.Index[T]` |

## Your Task

Ask the user what they're storing. Based on their answer, invoke the appropriate sub-skill:

- **Structured data with queries** (users, orders, documents) → `/add-store-database`
- **Files or blobs** (images, uploads, serialized objects) → `/add-store-bucket`
- **Simple key-value lookups** (cache, sessions, feature flags) → `/add-store-kv`
- **Embeddings for similarity search** (semantic search, recommendations) → `/add-store-index`

Do not proceed without invoking the appropriate sub-skill. Each variant has specific patterns and requirements.

## Store Location

Stores live in `stores/` with this structure:

```
stores/
├── stores.go      # Aggregates all stores, New() constructor
├── users.go       # Individual store implementation
├── documents.go
└── ...
```

Each store file contains:
1. The store struct with its grub wrapper
2. A constructor function
3. Methods that satisfy contracts

## Integration Points

- **grub** (`github.com/zoobzio/grub`) - Storage abstraction
- **soy** (`github.com/zoobzio/soy`) - SQL query builder (Database variant)
- **edamame** (`github.com/zoobzio/edamame`) - Statement executor (Database variant)
- **vecna** (`github.com/zoobzio/vecna`) - Vector filter builder (Index variant)
- **atom** (`github.com/zoobzio/atom`) - Atomization for type-agnostic access
- **scio** (`github.com/zoobzio/scio`) - URI-based data catalog

Review the GitHub repos if you need deeper understanding of any package.
