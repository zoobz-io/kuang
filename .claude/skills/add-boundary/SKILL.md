---
name: add-boundary
description: Add data transformation at system boundaries using cereal
---

# Add Boundary

You are adding boundary processing to a type - automatic field transformations that occur when data crosses system boundaries. Boundaries handle encryption, hashing, masking, and redaction based on struct tags.

## Technical Foundation

Boundaries are powered by `github.com/zoobzio/cereal` and integrated via `github.com/zoobzio/sum`. The processor transforms fields based on four contexts representing boundary crossings.

### The Four Contexts

| Context | Direction | When | Use Case |
|---------|-----------|------|----------|
| `receive` | External → App | API request arrives | Hash passwords |
| `load` | Storage → App | Read from database | Decrypt sensitive fields |
| `store` | App → Storage | Write to database | Encrypt sensitive fields |
| `send` | App → External | API response sent | Mask/redact for clients |

### Boundary Tags

**Receive context** (ingress from external):

```go
Password string `receive.hash:"argon2"`  // Hash with Argon2id
Password string `receive.hash:"bcrypt"`  // Hash with bcrypt
```

**Load context** (ingress from storage):

```go
AccessToken string `load.decrypt:"aes"`      // Decrypt with AES-GCM
PrivateKey  string `load.decrypt:"envelope"` // Decrypt with envelope encryption
```

**Store context** (egress to storage):

```go
AccessToken string `store.encrypt:"aes"`      // Encrypt with AES-GCM
PrivateKey  string `store.encrypt:"envelope"` // Encrypt with envelope encryption
```

**Send context** (egress to external):

```go
Email    string `send.mask:"email"`   // a]***@example.com
Phone    string `send.mask:"phone"`   // (***) ***-4567
SSN      string `send.mask:"ssn"`     // ***-**-6789
Card     string `send.mask:"card"`    // ************1111
IP       string `send.mask:"ip"`      // 192.168.xxx.xxx
Name     string `send.mask:"name"`    // J*** S****
Password string `send.redact:"***"`   // Complete replacement
APIKey   string `send.redact:"[REDACTED]"`
```

### Combining Tags

Fields can have multiple boundary tags:

```go
// Encrypted at rest, masked on send
Email string `store.encrypt:"aes" load.decrypt:"aes" send.mask:"email"`

// Hashed on receive, redacted on send
Password string `receive.hash:"argon2" send.redact:"***"`
```

### Wire Types (Request/Response)

For API boundary types in `wire/`, use `receive.*` and `send.*` tags.

**Response with masking:**

```go
type UserResponse struct {
    Email string `json:"email" send.mask:"email"`
    Phone string `json:"phone" send.mask:"phone"`
}

func (u *UserResponse) OnSend(ctx context.Context) error {
    b := sum.MustUse[*sum.Boundary[UserResponse]](ctx)
    masked, err := b.Send(ctx, *u)
    if err != nil {
        return err
    }
    *u = masked
    return nil
}

func (u UserResponse) Clone() UserResponse { return u }
```

**Request with hashing:**

```go
type RegisterRequest struct {
    Password string `json:"password" receive.hash:"argon2"`
}

func (r *RegisterRequest) OnEntry(ctx context.Context) error {
    b := sum.MustUse[*sum.Boundary[RegisterRequest]](ctx)
    processed, err := b.Receive(ctx, *r)
    if err != nil {
        return err
    }
    *r = processed
    return nil
}

func (r RegisterRequest) Clone() RegisterRequest { return r }
```

### Model Types (Database)

For domain models in `models/`, use `store.*` and `load.*` tags with lifecycle hooks.

**Model with encryption:**

```go
package models

import "context"

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

### Boundary Registration

Types using boundaries must be registered before `sum.Freeze()`:

```go
// wire/boundary.go
package wire

import "github.com/zoobzio/sum"

func RegisterBoundaries(k sum.Key) error {
    if _, err := sum.NewBoundary[UserResponse](k); err != nil {
        return err
    }
    if _, err := sum.NewBoundary[RegisterRequest](k); err != nil {
        return err
    }
    return nil
}
```

```go
// models/boundary.go
package models

import "github.com/zoobzio/sum"

func RegisterBoundaries(k sum.Key) error {
    if _, err := sum.NewBoundary[OAuthToken](k); err != nil {
        return err
    }
    return nil
}
```

### Encryption Setup

Encryption requires key configuration in the service:

```go
// In main.go or service setup
import (
    "github.com/zoobzio/cereal"
    "github.com/zoobzio/sum"
)

// Configure AES encryption
aesKey := []byte(cfg.EncryptionKey) // 32 bytes for AES-256
enc, err := cereal.AES(aesKey)
if err != nil {
    return err
}
sum.SetEncryptor(cereal.EncryptAES, enc)
```

### Available Algorithms

**Encryption (`store.encrypt` / `load.decrypt`):**

| Algorithm | Tag Value | Use Case |
|-----------|-----------|----------|
| AES-GCM | `aes` | Symmetric encryption for tokens, keys |
| RSA-OAEP | `rsa` | Asymmetric encryption |
| Envelope | `envelope` | Per-record data keys |

**Hashing (`receive.hash`):**

| Algorithm | Tag Value | Use Case |
|-----------|-----------|----------|
| Argon2id | `argon2` | Password hashing (recommended) |
| bcrypt | `bcrypt` | Password hashing (legacy) |
| SHA-256 | `sha256` | Deterministic hashing |
| SHA-512 | `sha512` | Deterministic hashing |

**Masking (`send.mask`):**

| Type | Tag Value | Example Output |
|------|-----------|----------------|
| Email | `email` | `a***@example.com` |
| Phone | `phone` | `(***) ***-4567` |
| SSN | `ssn` | `***-**-6789` |
| Card | `card` | `************1111` |
| IP | `ip` | `192.168.xxx.xxx` |
| UUID | `uuid` | `550e8400-****-****-****-************` |
| IBAN | `iban` | `GB82**************5432` |
| Name | `name` | `J*** S****` |

### Clone Requirement

All types using boundaries must implement `Clone() T`:

```go
// Simple value type - shallow copy is safe
func (u User) Clone() User { return u }

// Type with pointers/slices - must deep copy
func (o Order) Clone() Order {
    c := o
    if o.Items != nil {
        c.Items = make([]Item, len(o.Items))
        copy(c.Items, o.Items)
    }
    if o.Billing != nil {
        addr := *o.Billing
        c.Billing = &addr
    }
    return c
}
```

## Your Task

Understand what the user needs:

1. What type needs boundary processing?
2. Is it a wire type (request/response) or model (database)?
3. What fields need transformation?
4. What direction(s): receive, load, store, send?
5. What algorithms/mask types are needed?

## Before Writing Code

Produce a spec for approval:

```
## Boundary: [TypeName]

**Location:** [wire/domain.go or models/domain.go]

**Kind:** [Wire Request / Wire Response / Model]

**Fields with boundaries:**

[FieldName] ([type])
  [context].[action]:"[capability]"
  Purpose: [Why this transformation]

**Lifecycle method:** [OnEntry / OnSend / BeforeSave / AfterLoad]

**Encryption key required:** [yes/no - which algorithm]
```

## After Approval

1. Add boundary tags to the struct fields
2. Implement `Clone() T` method
3. Add appropriate lifecycle method:
   - Wire request: `OnEntry(ctx) error`
   - Wire response: `OnSend(ctx) error`
   - Model store: `BeforeSave(ctx) error`
   - Model load: `AfterLoad(ctx) error`
4. Register the boundary in `wire/boundary.go` or `models/boundary.go`
5. If encryption used: ensure key is configured in service setup
