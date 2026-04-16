---
name: add-event
description: Create events using capitan signals
---

# Add Event

You are creating events - signals that notify the system when something happens. Events are powered by `github.com/zoobzio/capitan` with typed wrappers from `github.com/zoobzio/sum`.

## Technical Foundation

Events live in `events/` organized by domain. There are two patterns depending on use case.

### Pattern 1: Typed Domain Events (sum.Event[T])

For domain lifecycle events that carry structured data:

```go
package events

import (
    "github.com/zoobzio/capitan"
    "github.com/zoobzio/sum"
)

// UserEvent carries user lifecycle data.
type UserEvent struct {
    UserID int64  `json:"user_id"`
    Login  string `json:"login"`
    Email  string `json:"email,omitempty"`
}

// User signals.
var (
    UserCreatedSignal  = capitan.NewSignal("app.user.created", "User created")
    UserUpdatedSignal  = capitan.NewSignal("app.user.updated", "User updated")
    UserLoggedInSignal = capitan.NewSignal("app.user.logged_in", "User logged in")
)

// User provides access to user lifecycle events.
var User = struct {
    Created  sum.Event[UserEvent]
    Updated  sum.Event[UserEvent]
    LoggedIn sum.Event[UserEvent]
}{
    Created:  sum.NewInfoEvent[UserEvent](UserCreatedSignal),
    Updated:  sum.NewInfoEvent[UserEvent](UserUpdatedSignal),
    LoggedIn: sum.NewInfoEvent[UserEvent](UserLoggedInSignal),
}
```

**Emitting:**
```go
events.User.Created.Emit(ctx, events.UserEvent{
    UserID: user.ID,
    Login:  user.Login,
    Email:  user.Email,
})
```

**Listening:**
```go
listener := events.User.Created.Listen(func(ctx context.Context, e events.UserEvent) {
    log.Printf("User created: %s", e.Login)
})
defer listener.Close()
```

### Pattern 2: Direct Signals (capitan)

For operational/infrastructure events without domain data:

```go
package events

import "github.com/zoobzio/capitan"

// Startup signals.
var (
    StartupDatabaseConnected = capitan.NewSignal("app.startup.database.connected", "Database connected")
    StartupServerListening   = capitan.NewSignal("app.startup.server.listening", "Server listening")
    StartupFailed            = capitan.NewSignal("app.startup.failed", "Startup failed")
)

// Field keys for direct emission.
var (
    StartupPortKey  = capitan.NewIntKey("port")
    StartupErrorKey = capitan.NewErrorKey("error")
)
```

**Emitting:**
```go
capitan.Info(ctx, events.StartupServerListening, events.StartupPortKey.Field(8080))
capitan.Error(ctx, events.StartupFailed, events.StartupErrorKey.Field(err))
```

### Severity Levels

- `sum.NewDebugEvent[T]` / `capitan.Debug` - development troubleshooting
- `sum.NewInfoEvent[T]` / `capitan.Info` - normal operations (default)
- `sum.NewWarnEvent[T]` / `capitan.Warn` - warning conditions
- `sum.NewErrorEvent[T]` / `capitan.Error` - error conditions

### Signal Naming Convention

```
{app}.{domain}.{action}
```

Examples:
- `app.user.created`
- `app.ingest.completed`
- `app.startup.database.connected`

### Field Key Types

For direct capitan usage:

```go
capitan.NewStringKey("name")
capitan.NewIntKey("count")
capitan.NewInt64Key("id")
capitan.NewFloat64Key("duration_ms")
capitan.NewBoolKey("success")
capitan.NewTimeKey("timestamp")
capitan.NewDurationKey("elapsed")
capitan.NewErrorKey("error")
```

## Your Task

Understand what events are needed:
1. What domain are these events for? (user, order, ingest, etc.)
2. What actions trigger events? (created, updated, completed, failed)
3. What data should events carry?
4. Are these domain events (typed) or operational events (direct)?

## Before Writing Code

Produce a spec for approval:

```
## Events: [Domain]

**Pattern:** [Typed (sum.Event) / Direct (capitan)]

**Event struct:** (if typed)
[FieldName] ([type]) - [purpose]

**Signals:**

[SignalName] ([severity])
  Name: [app.domain.action]
  Description: [Human-readable description]

**Field keys:** (if direct)
[KeyName] ([type]) - [purpose]
```

## After Approval

1. Create `events/[domain].go` with signals and event structs
2. Export signals for direct access when needed
3. For typed events, create the namespace struct with `sum.Event[T]` fields
4. Document usage in the package comment
