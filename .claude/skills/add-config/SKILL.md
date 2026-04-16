---
name: add-config
description: Create a new configuration type in the config/ package
---

# Add Config

You are creating a configuration type for this application. Before writing any code, you must understand what the user is building and why.

## Technical Foundation

Configuration types live in `config/` and are powered by `github.com/zoobzio/fig` for loading and `github.com/zoobzio/check` for validation. Review the GitHub repos if you need deeper understanding.

### Config Structure (zoobzio/fig)

```go
package config

import (
    "time"

    "github.com/zoobzio/check"
)

// TypeName holds configuration for [purpose].
type TypeName struct {
    Host     string        `env:"APP_SERVICE_HOST" default:"localhost"`
    Port     int           `env:"APP_SERVICE_PORT" default:"8080"`
    Timeout  time.Duration `env:"APP_SERVICE_TIMEOUT" default:"30s"`
    APIKey   string        `env:"APP_SERVICE_API_KEY" secret:"service/api-key"`
    Debug    bool          `env:"APP_SERVICE_DEBUG"`
    Replicas []string      `env:"APP_SERVICE_REPLICAS"` // comma-separated
}
```

**Struct tags (fig):**

- `env:"..."` - Environment variable name (you must ask the user, never assume)
- `default:"..."` - Default value if not set (optional)
- `secret:"..."` - Secret path for sensitive values (see Secrets section below)
- `required:"true"` - Fail if value is missing (use sparingly, prefer validation)

**Resolution order:** secret → env → default → zero value

**Supported types:** `string`, `int`, `float64`, `bool`, `time.Duration`, `[]string` (comma-separated)

**Avoid nested structs.** Each config should be flat and single-purpose. If related configuration feels like it belongs together (e.g., GitHub OAuth settings alongside app settings), create separate config files instead. This keeps configs independently loadable and testable.

### Validation (github.com/zoobzio/check)

Every config must have a `Validate() error` method using `github.com/zoobzio/check`:

```go
func (c TypeName) Validate() error {
    return check.All(
        // Required fields
        check.Required(c.Host, "host"),
        check.Required(c.APIKey, "api_key"),

        // Numeric constraints
        check.Positive(c.Port, "port"),
        check.Max(c.Port, 65535, "port"),

        // Duration constraints
        check.DurationPositive(c.Timeout, "timeout"),
        check.DurationMax(c.Timeout, 10*time.Minute, "timeout"),

        // String length
        check.MaxLen(c.Host, 255, "host"),
    ).Err()
}
```

**Fluent builders** for more complex validation:

```go
func (c TypeName) Validate() error {
    return check.All(
        // String builder - chain validations
        check.Str(c.Host, "host").Required().MaxLen(255).V(),

        // Optional string - only validates if non-empty
        check.OptStr(c.Prefix, "prefix").MaxLen(32).V(),

        // Integer builder
        check.Int(c.Port, "port").Positive().Max(65535).V(),

        // Duration builder
        check.Dur(c.Timeout, "timeout").Positive().Max(10*time.Minute).V(),
    ).Err()
}
```

**Common check functions:**

- `check.Required(v, name)` - value must be non-zero
- `check.Positive(n, name)` / `check.NonNegative(n, name)` - numeric bounds
- `check.Min(n, min, name)` / `check.Max(n, max, name)` - numeric limits
- `check.MinLen(s, n, name)` / `check.MaxLen(s, n, name)` - string length
- `check.DurationPositive(d, name)` / `check.DurationNonNegative(d, name)`
- `check.DurationMin(d, min, name)` / `check.DurationMax(d, max, name)`

### Helper Methods

Add helper methods when the user needs derived values:

```go
// DSN returns the PostgreSQL connection string.
func (c Database) DSN() string {
    return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode)
}

// Addr returns the host:port address.
func (c Server) Addr() string {
    return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
```

### Secrets

For sensitive values (passwords, API keys, tokens), use the `secret` tag instead of `env`:

```go
type Database struct {
    Host     string `env:"APP_DB_HOST" default:"localhost"`
    Password string `secret:"database/credentials:password"` // loaded from secret provider
}
```

The secret path format depends on the provider:
- **Vault**: `path/to/secret:field` or `path/to/secret` (defaults to "value" field)
- **AWS Secrets Manager**: `secret-name:field` or `secret-name`
- **GCP Secret Manager**: `secret-name`

Resolution order: secret → env → default. Secrets take precedence when a provider is configured.

**Important:** If the config needs secrets, the application must have a secret provider configured. If one doesn't exist yet, use the `/add-secret-manager` skill first to set it up.

### How Configs are Consumed

```go
// In main.go - loaded during startup (section 1)

// Without secrets - pass nil
if err := sum.Config[config.TypeName](ctx, k, nil); err != nil {
    return fmt.Errorf("failed to load typename config: %w", err)
}

// With secrets - pass the provider (must be initialized earlier in main.go)
if err := sum.Config[config.Database](ctx, k, secretProvider); err != nil {
    return fmt.Errorf("failed to load database config: %w", err)
}

// Retrieved where needed
cfg := sum.MustUse[config.TypeName](ctx)
```

## Your Task

Have a conversation with the user to understand:

- What they are configuring (a database? an external API? a service?)
- How the application will use this configuration
- What values need to be configurable

Do not ask these as a checklist. Understand the problem they're solving.

## Before Writing Code

You must produce a spec and get explicit approval. The spec demonstrates your understanding:

```
## Config: [Name]

**Purpose:** [What this configures and why]

**Used by:** [What part of the application consumes this]

**Fields:**

[FieldName] ([type])
  env: [exact env var]
  default: [value or "none"]
  secret: [vault path or "no"]

[NextField] ...

**Validation:**
[What makes this config invalid - expressed as check rules]

**Helper methods:**
[Any derived value methods, or "none"]
```

Wait for the user to approve before writing any files.

## After Approval

1. Create `config/[name].go` with the agreed spec
2. Update `cmd/app/main.go`:
   - Add the config import if needed
   - Add `sum.Config[config.TypeName](ctx, k, nil)` in the "1. Load Configuration" section
   - If config uses secrets, pass the secret provider instead of nil (if no provider exists, use `/add-secret-manager` first)
3. Update `.env.example` with the new environment variables under an appropriate section header
4. Update `docker-compose.yml` app service environment section for local dev
