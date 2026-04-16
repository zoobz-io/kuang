---
name: add-secret-manager
description: Configure a secret provider for loading sensitive configuration values
---

# Add Secret Manager

You are configuring a secret provider so the application can load sensitive values (passwords, API keys, tokens) from a secure backend instead of environment variables.

## Technical Foundation

Secret providers implement `fig.SecretProvider` and are passed to `sum.Config` when loading configurations. The provider is initialized once during startup and shared across all config loading.

### Available Providers (github.com/zoobzio/fig)

**HashiCorp Vault** (`github.com/zoobzio/fig/vault`):

```go
import "github.com/zoobzio/fig/vault"

// Uses VAULT_ADDR and VAULT_TOKEN from environment
provider, err := vault.New()

// Or with explicit options
provider, err := vault.New(
    vault.WithAddress("https://vault.example.com"),
    vault.WithToken(os.Getenv("VAULT_TOKEN")),
    vault.WithMount("secret"), // KV v2 mount path, default "secret"
)
```

Environment variables:
- `VAULT_ADDR` - Vault server address (required)
- `VAULT_TOKEN` - Authentication token (required for most auth methods)
- `VAULT_CACERT` - Path to CA cert for TLS (optional)

Secret key format: `path/to/secret:field` or `path/to/secret` (defaults to "value" field)

**AWS Secrets Manager** (`github.com/zoobzio/fig/awssm`):

```go
import "github.com/zoobzio/fig/awssm"

// Uses default AWS credential chain
provider, err := awssm.New(ctx)

// With explicit region
provider, err := awssm.New(ctx, awssm.WithRegion("us-east-1"))
```

Secret key format: `secret-name:field` or `secret-name` (parses JSON, uses field or entire value)

**GCP Secret Manager** (`github.com/zoobzio/fig/gcpsm`):

```go
import "github.com/zoobzio/fig/gcpsm"

// Uses default credentials and project from environment
provider, err := gcpsm.New(ctx)

// With explicit project
provider, err := gcpsm.New(ctx, gcpsm.WithProject("my-project"))
```

Secret key format: `secret-name` (uses latest version)

### Integration with sum.Config

```go
// In main.go - create provider once
provider, err := vault.New()
if err != nil {
    return fmt.Errorf("failed to create secret provider: %w", err)
}

// Pass to each config that needs secrets
if err := sum.Config[config.Database](ctx, k, provider); err != nil {
    return fmt.Errorf("failed to load database config: %w", err)
}

// Configs without secrets can pass nil
if err := sum.Config[config.Server](ctx, k, nil); err != nil {
    return fmt.Errorf("failed to load server config: %w", err)
}
```

### Using Secrets in Config Structs

```go
type Database struct {
    Host     string `env:"APP_DB_HOST" default:"localhost"`
    Port     int    `env:"APP_DB_PORT" default:"5432"`
    User     string `env:"APP_DB_USER"`
    Password string `secret:"database/credentials:password"` // loaded from secret provider
    Name     string `env:"APP_DB_NAME"`
}
```

Resolution order: secret → env → default → zero value

If a field has both `secret` and `env` tags, the secret takes precedence when available.

## Your Task

Understand what the user needs:
- Which secret backend they're using (Vault, AWS, GCP)
- How authentication will work in their environment
- Whether this is for local dev, production, or both

## Before Writing Code

Produce a spec for approval:

```
## Secret Manager: [Provider]

**Backend:** [Vault/AWS Secrets Manager/GCP Secret Manager]

**Authentication:** [How the app will authenticate - env vars, IAM role, etc.]

**Configuration:**
[List any env vars or options needed]

**Local dev setup:**
[How to run this locally - docker container, mock, etc.]
```

## After Approval

1. Update `cmd/app/main.go`:
   - Add the provider import
   - Initialize the provider in section 1 (before config loading)
   - Pass the provider to configs that need secrets
2. Update `.env.example` with required environment variables
3. Update `docker-compose.yml` if a local secret backend is needed (e.g., Vault dev server)
