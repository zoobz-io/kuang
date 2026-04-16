---
name: add-pipeline
description: Build composable data processing workflows using pipz
---

# Add Pipeline

You are building a pipeline using `github.com/zoobzio/pipz`. Pipelines enable clean, testable, and observable execution of complex processing workflows.

**This skill teaches patterns and the pipz abstraction, not structure.** The pipeline you build should reflect the problem being solved. There is no prescribed stage count, naming convention, or flow - design what the user's problem requires.

## The Core Abstraction

Everything in pipz implements one interface:

```go
type Chainable[T any] interface {
    Process(context.Context, T) (T, error)
    Identity() Identity
    Schema() Node
    Close() error
}
```

This is the **fundamental insight**: any type implementing `Chainable[T]` becomes a first-class primitive that can be used anywhere in pipz. Adapters, connectors, reliability wrappers - they all implement this interface. A `Sequence` is a `Chainable`. A `Retry` wrapping a `Chainable` is itself a `Chainable`. This enables infinite composition.

**The constraint enables expression:** Because everything shares this interface, you can compose any Chainable with any other, in any order, to any depth, and the result is always another `Chainable[T]` that can be composed further.

## The Carrier Type

**Critical: T is fixed for the entire pipeline.** You cannot transform from `T` to `K` within a pipeline - the same type flows through every stage. This constraint is by design: it enables serialization, resumability, and state tracking.

A **carrier** is the recommended pattern for T. Design a type that:
- Holds everything stages need to do their work
- Accumulates state as processing progresses
- Can be serialized and stored (for persistence, debugging, replay)
- Can be cloned for parallel execution

```go
// Job is the carrier - it flows through the entire pipeline.
type Job struct {
    // Identity - what is being processed
    ID           int64
    UserID       int64
    TargetID     int64

    // Progress tracking
    Stage        string
    Status       string
    Progress     int
    ItemsTotal   int
    ItemsProcessed int

    // Error state
    Error        *string

    // Timestamps
    CreatedAt    time.Time
    StartedAt    *time.Time
    CompletedAt  *time.Time
}

// Clone implements pipz.Cloner - required for parallel execution.
func (j *Job) Clone() *Job {
    c := *j
    if j.Error != nil {
        e := *j.Error
        c.Error = &e
    }
    return &c
}
```

The carrier is YOUR design - it should reflect what your pipeline processes. If a stage needs to fetch data, the carrier holds the result. If a stage needs configuration, the carrier provides it.

## Stages

Pipelines are naturally stage-based. Each stage is an atomic unit of work - focused, testable, and composable. **What stages you need emerges from the problem**, but the staged approach itself is the pattern:

- **Atomic** - Each stage does one thing well
- **Testable** - Stages can be tested in isolation
- **Composable** - Stages combine via pipz connectors
- **Observable** - Each stage has identity for debugging

A stage is just a function wrapped in an adapter:

```go
var FetchID = pipz.NewIdentity("fetch", "Fetches data from source")

func fetchStage(ctx context.Context, job *Job) (*Job, error) {
    service := sum.MustUse[contracts.Source](ctx)

    data, err := service.Fetch(ctx, job.TargetID)
    if err != nil {
        return job, err
    }

    job.Data = data
    job.Stage = "fetch"
    return job, nil
}

// Wrap as a Chainable
fetch := pipz.Apply(FetchID, fetchStage)
```

## Processors and Connectors

pipz distinguishes between two types of Chainables:

**Processors** (adapters that wrap stage functions):
- Lightweight, stateless, immutable after creation
- Wrap your functions to implement Chainable
- `Apply`, `Transform`, `Effect`, `Mutate`, `Enrich`

**Connectors** (orchestrate stages):
- Manage execution flow between stages
- Accept any Chainable - processors or other connectors
- `Sequence`, `Concurrent`, `Race`, `Switch`, `Retry`, `Timeout`, etc.

Stages contain the logic. Connectors compose stages into pipelines.

## Composing Pipelines

### Sequential Processing

```go
pipeline := pipz.NewSequence(ID,
    pipz.Apply(ValidateID, validate),
    pipz.Apply(ProcessID, process),
    pipz.Apply(FinalizeID, finalize),
)
```

### Resilience Wrapping

Because resilience wrappers accept any Chainable, they compose naturally:

```go
protected := pipz.NewTimeout(TimeoutID,
    pipz.NewRetry(RetryID,
        pipz.NewFallback(FallbackID, primary, backup),
        3),
    5*time.Minute)
```

This reads like prose: timeout a retry of a fallback. The result is a single `Chainable[T]` usable anywhere.

### Parallel Execution

```go
race := pipz.NewRace(ID,
    fetchFromCache,
    fetchFromDB,
    fetchFromRemote)
```

First successful result wins. All three are Chainables - simple Apply functions or complex nested pipelines.

### Conditional Routing

```go
router := pipz.NewSwitch(ID, func(ctx context.Context, job *Job) string {
    if job.Priority == "high" { return "fast" }
    return "standard"
})
router.AddRoute("fast", fastPipeline)
router.AddRoute("standard", standardPipeline)
```

### Error as Data

Errors flow through pipelines as first-class values:

```go
errorPipeline := pipz.NewSequence[*pipz.Error[*Job]](ID,
    logError,
    classifyError,
    recoveryRouter)
mainPipeline := pipz.NewHandle(ID, processing, errorPipeline)
```

The `Error[T]` type carries rich context: path through pipeline, input data at failure, duration, timeout/cancellation status.

### Complete Example

```go
func NewPipeline() *pipz.Sequence[*Job] {
    validate := pipz.Apply(ValidateID, validateStage)
    process := pipz.Apply(ProcessID, processStage)
    finalize := pipz.Apply(FinalizeID, finalizeStage)

    // Compose reliability around the processing stage
    processReliable := pipz.NewTimeout(TimeoutID,
        pipz.NewRetry(RetryID, process, 3),
        5*time.Minute)

    return pipz.NewSequence(PipelineID,
        validate,
        processReliable,
        finalize,
    )
}
```

### The Worker

A worker manages the pipeline lifecycle - triggering execution, tracking state, emitting events:

```go
type Worker struct {
    pool     *pipz.WorkerPool[*Job]
    pipeline *pipz.Sequence[*Job]
    listener *capitan.Listener
}

func NewWorker() *Worker {
    pipeline := NewPipeline()
    return &Worker{
        pool:     pipz.NewWorkerPool(WorkerPoolID, 4, pipeline),
        pipeline: pipeline,
    }
}

func (w *Worker) Start(ctx context.Context) {
    w.listener = events.Job.Created.Listen(func(ctx context.Context, e events.JobCreatedEvent) {
        w.process(ctx, e.Job)
    })
}

func (w *Worker) process(ctx context.Context, job *Job) {
    jobs := sum.MustUse[contracts.Jobs](ctx)

    // Mark started
    jobs.Start(ctx, job.ID)
    events.Job.Started.Emit(ctx, ...)

    // Execute pipeline
    result, err := w.pool.Process(ctx, job)

    if err != nil {
        jobs.MarkFailed(ctx, job.ID, err.Error())
        events.Job.Failed.Emit(ctx, ...)
        return
    }

    jobs.MarkCompleted(ctx, result.ID)
    events.Job.Completed.Emit(ctx, ...)
}

func (w *Worker) Stop() error {
    if w.listener != nil {
        w.listener.Close()
    }
    return w.pool.Close()
}
```

### Writing Stage Functions

Stage functions process the carrier. They should be focused and testable:

```go
func processStage(ctx context.Context, job *Job) (*Job, error) {
    // Resolve dependencies
    service := sum.MustUse[contracts.Service](ctx)

    // Do one thing
    results, err := service.Process(ctx, job.TargetID)
    if err != nil {
        return job, err
    }

    // Update carrier state
    job.ItemsProcessed = len(results)
    job.Stage = "process"

    return job, nil
}
```

Keep stages atomic. If a stage is doing multiple distinct operations, it's probably multiple stages.

### Nested Pipelines (Different Carriers)

Remember: T is fixed **per pipeline**. When a stage needs to do parallel sub-work with different data, create a separate pipeline with its own carrier type. This is the same abstraction - just a different T:

```go
// fileWork is a different carrier for a different pipeline
type fileWork struct {
    JobID   int64
    Path    string
    Content []byte
}

func (w *fileWork) Clone() *fileWork {
    c := *w
    return &c
}

// A separate pipeline for file processing
var filePool = pipz.NewWorkerPool(FilePoolID, 8,
    pipz.Apply(ProcessFileID, processFile),
).WithTimeout(30 * time.Second)

// Stage in the main pipeline dispatches to the nested pipeline
func fetchStage(ctx context.Context, job *Job) (*Job, error) {
    for _, file := range files {
        go filePool.Process(ctx, &fileWork{
            JobID:   job.ID,
            Path:    file.Path,
            Content: file.Content,
        })
    }
    return job, nil
}
```

This is not a special feature - it's the same Chainable abstraction with a different type parameter.

## The pipz Vocabulary

All of these implement `Chainable[T]`. They are the vocabulary for expressing your pipeline.

### Adapters (wrap your functions)

| Adapter | Signature | Use Case |
|---------|-----------|----------|
| `Apply` | `func(ctx, T) (T, error)` | Operations that can fail |
| `Transform` | `func(ctx, T) T` | Pure transformations |
| `Effect` | `func(ctx, T) error` | Side effects without modifying data |
| `Mutate` | `func + predicate` | Conditional modifications |
| `Enrich` | `func(ctx, T) (T, error)` | Optional enhancements (logs failures, continues) |

### Connectors (compose Chainables)

| Connector | Purpose |
|-----------|---------|
| `Sequence` | Chain sequentially |
| `Concurrent` | Run all in parallel, return original |
| `Race` | Return first successful result |
| `Contest` | Return first meeting a condition |
| `Switch` | Route based on conditions |
| `Filter` | Conditional processing |

### Reliability (wrap any Chainable)

| Wrapper | Purpose |
|---------|---------|
| `Timeout` | Enforce time limits |
| `Retry` | Retry N times on failure |
| `Backoff` | Retry with exponential backoff |
| `Fallback` | Try backup on error |
| `Handle` | Route errors to a pipeline |
| `CircuitBreaker` | Prevent cascade failures |
| `RateLimiter` | Control throughput |

### Execution

| Component | Purpose |
|-----------|---------|
| `WorkerPool` | Bounded concurrent execution |
| `Scaffold` | Ordered parallel processing |
| `Pipeline` | Adds execution context and signals |

### Identity

Every component needs an identity for observability and debugging:

```go
var StageID = pipz.NewIdentity("stage-name", "Human description")
```

Identities enable pipeline visualization, runtime inspection, and correlating errors to their source.

## Your Task

Understand what the user is building:

1. **What is the goal?** What should this pipeline accomplish?
2. **What is the carrier?** What data needs to flow through and accumulate?
3. **What stages are needed?** What atomic units of work does the problem require? Stages emerge from the problem - identify them, don't invent them.
4. **How should stages be composed?** Sequential? Parallel? Conditional routing?
5. **What reliability is required?** Timeouts? Retries? Fallbacks?
6. **How is it triggered?** Event-driven? HTTP endpoint? Scheduled?

## Before Writing Code

Produce a spec for approval:

```
## Pipeline: [Name]

**Purpose:** [What this pipeline accomplishes]

**Carrier:** [What flows through - the state it tracks]

**Stages:** [What atomic units of work are needed - these emerge from the problem]
- [StageName]: [What this stage does]
- [StageName]: [What this stage does]
...

**Composition:** [How stages are connected - sequential, parallel, conditional, with what reliability wrappers]

**Trigger:** [How execution begins - event, endpoint, schedule]

**Execution:** [WorkerPool with concurrency, or direct Process call]
```

Stages emerge from the problem. The spec should identify the atomic units of work needed, then describe how pipz composes them.

## After Approval

1. Create the carrier type with `Clone()` method (see `/add-model`)
2. Create stage functions - one per atomic unit of work
3. Create the pipeline construction function that composes stages
4. Create any nested pipelines with their own carriers if needed
5. Create the execution wrapper (worker, handler, etc.)
6. Create events as needed (see `/add-event`)
7. Create contracts for dependencies (see `/add-contract`)
