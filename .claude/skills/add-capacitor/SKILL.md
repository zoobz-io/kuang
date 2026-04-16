---
name: add-capacitor
description: Create hot-reload runtime configuration using flux capacitors
---

# Add Capacitor

You are creating a capacitor - a hot-reload runtime configuration that can be changed without restarting the application. Capacitors watch configuration sources and apply changes in real-time.

**Use capacitors for operational settings that admins need to tune at runtime.** Use static config (see `/add-config`) for bootstrap settings that require restart.

## Technical Foundation

Capacitors are powered by `github.com/zoobzio/flux`.

### The Core Abstraction

```go
type Capacitor[T Validator] struct { ... }

type Watcher interface {
    Watch(ctx context.Context) (<-chan []byte, error)
}

type Validator interface {
    Validate() error
}
```

Any type with a `Validate()` method can be a capacitor config. Any source implementing `Watcher` can provide values.

### Basic Usage

```go
capacitor := flux.New[Config](
    watcher,
    func(ctx context.Context, prev, curr Config) error {
        // Apply the new config
        return nil
    },
)

if err := capacitor.Start(ctx); err != nil {
    // Initial load failed
}

// Access current config
if cfg, ok := capacitor.Current(); ok {
    // Use cfg
}
```

## Watcher Providers

### File (Local Filesystem)

```go
import "github.com/zoobzio/flux/file"

capacitor := flux.New[Config](
    file.New("/etc/myapp/config.json"),
    callback,
)
```

Best for: Local development, single-instance deployments.

### Redis

```go
import "github.com/zoobzio/flux/redis"

capacitor := flux.New[Config](
    redis.New(client, "myapp:config"),
    callback,
)
```

Requires: `CONFIG SET notify-keyspace-events KEA`

Best for: Feature flags, shared config across instances.

### Consul KV

```go
import "github.com/zoobzio/flux/consul"

capacitor := flux.New[Config](
    consul.New(client, "myapp/config"),
    callback,
)
```

Best for: HashiCorp ecosystem, service mesh configurations.

### etcd

```go
import "github.com/zoobzio/flux/etcd"

capacitor := flux.New[Config](
    etcd.New(client, "/myapp/config"),
    callback,
)
```

Best for: Kubernetes ecosystem, strong consistency requirements.

### NATS JetStream KV

```go
import "github.com/zoobzio/flux/nats"

capacitor := flux.New[Config](
    nats.New(kv, "myapp"),
    callback,
)
```

Best for: Cloud-native apps, microservices using NATS.

### Kubernetes ConfigMap/Secret

```go
import "github.com/zoobzio/flux/kubernetes"

// ConfigMap
capacitor := flux.New[Config](
    kubernetes.New(client, "default", "myapp-config", "config.json"),
    callback,
)

// Secret
capacitor := flux.New[Config](
    kubernetes.New(client, "default", "myapp-secret", "config.json",
        kubernetes.WithResourceType(kubernetes.Secret),
    ),
    callback,
)
```

Best for: Kubernetes-native applications, GitOps workflows.

### PostgreSQL (Custom Watcher)

For database-backed config with LISTEN/NOTIFY, implement a custom watcher (see vicky's `capacitors/watcher.go` for reference).

## The Config Struct

Define the configuration with JSON tags and validation:

```go
package capacitors

import (
    "time"

    "github.com/zoobzio/check"
)

// Fetch holds operational settings for the fetch stage.
type Fetch struct {
    Workers int           `json:"workers"` // pool concurrency
    Timeout time.Duration `json:"timeout"` // per-operation timeout
}

// Validate checks configuration constraints.
func (c Fetch) Validate() error {
    return check.All(
        check.NonNegative(c.Workers, "workers"),
        check.Max(c.Workers, 100, "workers"),
        check.DurationNonNegative(c.Timeout, "timeout"),
        check.DurationMax(c.Timeout, 10*time.Minute, "timeout"),
    ).Err()
}
```

Zero values can mean "use default" - design your validation accordingly.

### Default Function

```go
// DefaultFetch returns sensible defaults.
func DefaultFetch() Fetch {
    return Fetch{
        Workers: 8,
        Timeout: 30 * time.Second,
    }
}
```

## The Apply Pattern

The callback receives previous and current config:

```go
func(ctx context.Context, prev, curr Config) error {
    // Apply changes to running system
    ingest.SetFetchConfig(curr.Workers, curr.Timeout)
    return nil
}
```

The apply function bridges the capacitor to whatever it controls - worker pools, rate limiters, feature flags, etc.

## Initialization Pattern

```go
import (
    "context"
    "log"

    "github.com/zoobzio/flux"
)

// InitFetch initializes the fetch capacitor.
func InitFetch(ctx context.Context, watcher flux.Watcher) error {
    // Apply defaults immediately
    applyFetch(DefaultFetch())

    c := flux.New[Fetch](
        watcher,
        func(_ context.Context, _, curr Fetch) error {
            applyFetch(curr)
            return nil
        },
    )

    go func() {
        if err := c.Start(ctx); err != nil {
            log.Printf("fetch capacitor error: %v", err)
        }
    }()

    return nil
}
```

## Capacitor Options

### Debouncing

Prevent thrashing on rapid changes:

```go
capacitor := flux.New[Config](watcher, callback).
    Debounce(500 * time.Millisecond)  // Default: 100ms
```

### Codec

Use YAML instead of JSON:

```go
capacitor := flux.New[Config](watcher, callback).
    Codec(flux.YAMLCodec{})
```

### Startup Timeout

Fail fast if initial config doesn't arrive:

```go
capacitor := flux.New[Config](watcher, callback).
    StartupTimeout(5 * time.Second)
```

## Reliability Patterns

flux integrates with pipz for resilience:

```go
capacitor := flux.New[Config](
    watcher,
    callback,
    flux.WithRetry[Config](3),
    flux.WithBackoff[Config](3, 100*time.Millisecond),
    flux.WithTimeout[Config](5*time.Second),
    flux.WithCircuitBreaker[Config](5, 30*time.Second),
)
```

## State Machine

Capacitors track state:

```go
const (
    StateLoading  // Waiting for first value
    StateHealthy  // Valid config active
    StateDegraded // Update failed, previous retained
    StateEmpty    // No valid config ever obtained
)

// Check state
state := capacitor.State()
```

- **Loading → Healthy**: First config loaded successfully
- **Loading → Empty**: First config failed validation
- **Healthy → Degraded**: Update failed, previous config retained
- **Degraded → Healthy**: Valid config arrived, recovered

## Multi-Source Composition

Merge multiple sources with a reducer:

```go
capacitor := flux.Compose[Config](
    func(ctx context.Context, prev, curr []Config) (Config, error) {
        // curr[0]=defaults, curr[1]=file, curr[2]=remote
        merged := curr[0]
        if curr[1].Workers != 0 {
            merged.Workers = curr[1].Workers
        }
        if curr[2].Workers != 0 {
            merged.Workers = curr[2].Workers
        }
        return merged, nil
    },
    []flux.Watcher{
        file.New("/etc/defaults.json"),
        file.New("/etc/config.json"),
        consul.New(client, "myapp/config"),
    },
)
```

## Observability

flux emits capitan signals:

- `CapacitorStarted` / `CapacitorStopped`
- `CapacitorStateChanged`
- `CapacitorChangeReceived`
- `CapacitorValidationFailed`
- `CapacitorApplyFailed` / `CapacitorApplySucceeded`

```go
capitan.Hook(flux.CapacitorStateChanged, func(ctx context.Context, e *capitan.Event) {
    old, _ := flux.KeyOldState.From(e)
    new, _ := flux.KeyNewState.From(e)
    log.Printf("state: %s -> %s", old, new)
})
```

## Common Patterns

### Pipeline Stage Capacitor

```go
type Stage struct {
    Workers int           `json:"workers"`
    Timeout time.Duration `json:"timeout"`
}
```

### Feature Flag Capacitor

```go
type Features struct {
    EnableBetaSearch  bool `json:"enable_beta_search"`
    EnableCaching     bool `json:"enable_caching"`
    MaxResultsPerPage int  `json:"max_results_per_page"`
}
```

### Rate Limit Capacitor

```go
type RateLimits struct {
    RequestsPerSecond int `json:"requests_per_second"`
    BurstSize         int `json:"burst_size"`
}
```

## Your Task

Understand what the user needs:

1. **What is being controlled?** Pipeline stage? Feature flags? Rate limits?
2. **What watcher fits?** File? Redis? Consul? Kubernetes?
3. **What fields are needed?** What values can admins tune?
4. **What are sensible defaults?** What should apply before any config arrives?
5. **What validation is needed?** What constraints keep the system stable?
6. **What reliability is needed?** Retries? Circuit breaker?

## Before Writing Code

Produce a spec for approval:

```
## Capacitor: [Name]

**Purpose:** [What this capacitor controls]

**Watcher:** [file/redis/consul/etcd/kubernetes/custom]

**Fields:**
- [field] ([type]): [purpose] (default: [value])
- ...

**Validation:**
- [constraint description]
- ...

**Apply:** [What system/component receives the config changes]

**Options:** [debounce/retry/circuit breaker if needed]
```

## After Approval

1. Create `capacitors/[name].go` with:
   - Config struct with `Validate()`
   - Default function
   - Apply function
   - Init function
2. Add initialization to `capacitors/capacitors.go`
3. If using file/consul/etc., set up the config source
4. If using database, create migration with LISTEN/NOTIFY trigger
5. Implement the target's config setter (e.g., `ingest.SetFetchConfig`)
