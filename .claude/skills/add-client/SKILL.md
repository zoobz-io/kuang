---
name: add-client
description: Create an external API client with resilience patterns
---

# Add Client

You are creating an external client - a wrapper around an HTTP API, gRPC service, or third-party SDK. All clients must implement resilience patterns using `github.com/zoobzio/pipz` to handle failures gracefully.

## Technical Foundation

Clients live in `external/[service]/client.go` and implement contracts from `contracts/`.

### Resilience Stack

Every external call should be wrapped in a resilience pipeline:

```
RateLimiter (if API has rate limits)
  → CircuitBreaker (prevent cascade failures)
    → Backoff (retry with exponential delay)
      → Timeout (prevent hanging)
        → Processor (the actual API call)
```

### Pattern 1: Direct pipz Usage

For HTTP/gRPC clients where you need full control:

```go
package github

import (
    "context"
    "time"

    "github.com/zoobzio/pipz"
)

// Resilience configuration.
const (
    apiTimeout          = 30 * time.Second
    apiMaxAttempts      = 3
    apiBackoffDelay     = 200 * time.Millisecond
    apiFailureThreshold = 5
    apiResetTimeout     = 30 * time.Second
    apiRatePerSecond    = 1.0
    apiRateBurst        = 10
)

// Pipeline identities.
var (
    callProcessorID   = pipz.NewIdentity("github.call", "GitHub API call")
    callTimeoutID     = pipz.NewIdentity("github.timeout", "Timeout for API calls")
    callBackoffID     = pipz.NewIdentity("github.backoff", "Backoff retry for API calls")
    callBreakerID     = pipz.NewIdentity("github.breaker", "Circuit breaker for API")
    callRateLimiterID = pipz.NewIdentity("github.ratelimit", "Rate limiter for API")
)

// apiCall carries request and response through the pipeline.
type apiCall struct {
    token    string
    endpoint string
    result   *Response
}

// Clone enables concurrent processing (required by pipz).
func (c *apiCall) Clone() *apiCall {
    clone := *c
    if c.result != nil {
        r := *c.result
        clone.result = &r
    }
    return &clone
}

// Client implements contracts.GitHub.
type Client struct {
    pipeline pipz.Chainable[*apiCall]
}

// NewClient creates a new API client with resilience.
func NewClient() *Client {
    c := &Client{}
    c.pipeline = c.buildPipeline()
    return c
}

// buildPipeline constructs the resilient processing pipeline.
func (c *Client) buildPipeline() pipz.Chainable[*apiCall] {
    processor := pipz.Apply(callProcessorID, func(ctx context.Context, call *apiCall) (*apiCall, error) {
        // Make the actual API call here
        // call.result = ...
        return call, nil
    })

    return pipz.NewRateLimiter(callRateLimiterID, apiRatePerSecond, apiRateBurst,
        pipz.NewCircuitBreaker(callBreakerID,
            pipz.NewBackoff(callBackoffID,
                pipz.NewTimeout(callTimeoutID, processor, apiTimeout),
                apiMaxAttempts, apiBackoffDelay,
            ),
            apiFailureThreshold, apiResetTimeout,
        ),
    )
}

// Close shuts down the pipeline.
func (c *Client) Close() error {
    if c.pipeline != nil {
        return c.pipeline.Close()
    }
    return nil
}
```

### Pattern 2: Library with Built-in Resilience

When using a library that wraps pipz internally (like vex for embeddings):

```go
package embedding

import (
    "time"

    "github.com/zoobzio/vex"
    "github.com/zoobzio/vex/openai"
)

// Client implements contracts.Embedder.
type Client struct {
    svc  *vex.Service
    dims int
}

// NewClient creates a new embedding client.
func NewClient(apiKey, model string, dimensions int) (*Client, error) {
    provider := openai.New(openai.Config{
        APIKey:     apiKey,
        Model:      model,
        Dimensions: dimensions,
    })

    svc := vex.NewService(provider,
        vex.WithBackoff(3, 100*time.Millisecond),
        vex.WithTimeout(30*time.Second),
        vex.WithCircuitBreaker(5, 30*time.Second),
    )

    return &Client{svc: svc, dims: dimensions}, nil
}
```

### Key Requirements

**Call types must implement Clone():**

```go
func (c *apiCall) Clone() *apiCall {
    clone := *c
    // Deep copy any slices, maps, or pointers
    if c.data != nil {
        clone.data = make([]byte, len(c.data))
        copy(clone.data, c.data)
    }
    return &clone
}
```

**Every component needs an Identity:**

```go
var (
    processorID = pipz.NewIdentity("service.operation.call", "Description")
    timeoutID   = pipz.NewIdentity("service.operation.timeout", "Description")
    backoffID   = pipz.NewIdentity("service.operation.backoff", "Description")
    breakerID   = pipz.NewIdentity("service.operation.breaker", "Description")
)
```

**Clients must implement Close():**

```go
func (c *Client) Close() error {
    if c.pipeline != nil {
        return c.pipeline.Close()
    }
    return nil
}
```

### Resilience Configuration Guidelines

| Parameter         | Typical Value | Notes                                        |
| ----------------- | ------------- | -------------------------------------------- |
| Timeout           | 30s           | Longer for heavy operations (indexing: 5min) |
| Max attempts      | 3             | Fewer for expensive operations               |
| Backoff delay     | 100-500ms     | Initial delay, grows exponentially           |
| Failure threshold | 5             | Failures before circuit opens                |
| Reset timeout     | 30s           | Time before circuit tries again              |
| Rate limit        | varies        | Match API's documented limits                |

## Your Task

Understand what the user is building:

1. What external service/API are they calling?
2. Does the API have rate limits?
3. What operations does the client need to support?
4. Is there an existing library with built-in resilience, or do we need direct pipz?

## Before Writing Code

Produce a spec for approval:

```
## Client: [Name]

**Service:** [External API/service being wrapped]

**Pattern:** [Direct pipz / Library with resilience]

**Package:** external/[name]/client.go

**Contract:** contracts.[Name]

**Operations:**

[MethodName](ctx, [params]) ([returns], error)
  Purpose: [What this operation does]
  Timeout: [Expected timeout]

**Resilience:**

Timeout: [duration]
Retries: [count] with [delay] backoff
Circuit breaker: [threshold] failures, [reset] timeout
Rate limit: [rate/sec, burst] (if applicable)
```

## After Approval

1. Create `external/[name]/client.go` with:
   - Call type(s) with Clone() method
   - Pipeline identity constants
   - Resilience configuration constants
   - Client struct with pipeline field
   - Constructor that builds the pipeline
   - Methods that use the pipeline
   - Close() method
2. Create `external/[name]/types.go` if needed for request/response types
3. Create or update `contracts/[name].go` with the interface
4. Register the client in `cmd/app/main.go`
