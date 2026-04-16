# Mission: sumatra

Application template for zoobzio Go services.

## Purpose

Provide a complete, ready-to-use foundation for new Go applications. Clone it, rename it, start building.

## The Stack

| Package | Purpose |
|---------|---------|
| `sum` | Service registry, dependency injection, boundaries |
| `rocco` | HTTP handlers, OpenAPI generation, SSE streaming |
| `grub` | Storage abstraction (Database, Bucket, Store, Index) |
| `soy` | Type-safe SQL query builder |
| `pipz` | Composable pipeline workflows |
| `flux` | Hot-reload runtime configuration (capacitors) |
| `cereal` | Field-level encryption, hashing, masking |
| `capitan` | Events and observability signals |
| `check` | Request validation |

## API Surfaces

This template supports multiple API surfaces — separate binaries serving different consumers.

| Surface | Binary | Consumer | Characteristics |
|---------|--------|----------|-----------------|
| `api` | `cmd/app/` | Customers | User-scoped, masked data, conservative exposure |
| `admin` | `cmd/admin/` | Internal team | System-wide, full visibility, bulk operations |

### Layer Organization

**Shared layers** (used by all surfaces):
- `models/` — Domain models
- `stores/` — Data access implementations (same store satisfies multiple contracts)
- `migrations/` — Database schema
- `events/` — Domain events
- `config/` — Configuration

**Surface-specific layers** (each surface has its own):
- `{surface}/contracts/` — Interface definitions
- `{surface}/wire/` — Request/response types (different masking per surface)
- `{surface}/handlers/` — HTTP handlers
- `{surface}/transformers/` — Model <-> wire mapping

### Surface Differences

| Aspect | Public (api/) | Admin (admin/) |
|--------|---------------|----------------|
| Auth | Customer identity | Admin/internal identity |
| Scope | User's own data | System-wide access |
| Operations | Standard CRUD | Bulk ops, impersonation, audit |
| Data exposure | Masked, minimal | Full visibility |

Note: Stores are shared. The same store implementation satisfies both `api/contracts.Users` and `admin/contracts.Users`.

## Project Structure

```
cmd/
├── app/              # Public API binary
└── admin/            # Admin API binary

# Shared layers
config/               # Static configuration types
models/               # Domain models
stores/               # Data access (shared, satisfies multiple contracts)
events/               # Domain events and signals
migrations/           # Database migrations (goose)
internal/             # Internal packages
testing/              # Test infrastructure, mocks, fixtures

# Public API surface
api/
├── contracts/        # Public interface definitions
├── wire/             # Public request/response types (masked)
├── handlers/         # Public HTTP handlers
└── transformers/     # Public model <-> wire mapping

# Admin API surface
admin/
├── contracts/        # Admin interface definitions
├── wire/             # Admin request/response types (unmasked)
├── handlers/         # Admin HTTP handlers
└── transformers/     # Admin model <-> wire mapping
```

## Conventions

### Naming

| Layer | File | Type | Example |
|-------|------|------|---------|
| Model | `models/user.go` | `User` (singular) | `type User struct` |
| Store | `stores/users.go` | `Users` (plural struct) | `type Users struct` |
| Contract | `{surface}/contracts/users.go` | `Users` (plural interface) | `type Users interface` |
| Wire | `{surface}/wire/users.go` | Singular + suffix | `UserResponse`, `AdminUserResponse` |
| Handler | `{surface}/handlers/users.go` | Verb+Singular | `var GetUser`, `var CreateUser` |

### Registration Points

After creating artifacts, wire them appropriately:

**Shared:**
- `stores/stores.go` — aggregate factory (all stores)
- `models/boundary.go` — model boundaries

**Surface-specific (replace `{surface}` with `api` or `admin`):**
- `{surface}/handlers/handlers.go` — `All()` function
- `{surface}/handlers/errors.go` — domain errors
- `{surface}/wire/boundary.go` — wire boundaries (masking for public API)

### Testing

- 1:1 relationship: `user.go` -> `user_test.go`
- Helpers in `testing/` call `t.Helper()`
- Mocks use function-field pattern
- Fixtures return test data with sensible defaults

## Success Criteria

A developer can:
1. Create a new repo from this template
2. Update MISSION.md and go.mod with their service name
3. Run `make check` successfully
4. Define their first entity using skills
5. Start building their application immediately

## Non-Goals

- Library infrastructure (that's samoa)
- Opinionated business patterns
- Pre-configured cloud integrations
