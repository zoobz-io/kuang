# Architecture Review Checklist

## Phase 1: Application Inventory

### Surfaces
- [ ] Inventory API surfaces (api/, admin/)
- [ ] Identify each surface's consumers and purpose

### Shared Layers
- [ ] Inventory shared layers (models/, stores/, migrations/, events/, config/)
- [ ] Verify shared layers serve all surfaces

### Surface-Specific Layers
- [ ] Inventory surface-specific layers per surface (contracts/, wire/, handlers/, transformers/)
- [ ] Verify each surface has its own complete set

### Internal Packages
- [ ] Inventory internal packages (internal/)

### Binaries
- [ ] Inventory binaries (cmd/)
- [ ] Identify each binary's purpose

### File Organization
- [ ] List all `.go` files and their apparent purpose
- [ ] Identify core files vs provider files vs utility files
- [ ] Files that serve multiple unrelated purposes are findings

## Phase 2: Interface Assessment — Where Is the Surface Too Wide?

### Width
- [ ] Does any interface have more methods than its consumers use? (unnecessary commitment)
- [ ] Are there "god interfaces" that force implementers to provide everything? (wide attack surface)
- [ ] Can interfaces be split into smaller, focused units?

### Composability
- [ ] Do larger interfaces embed smaller ones — or are they monolithic?
- [ ] Can interfaces be used independently — or only as parts of a forced whole?
- [ ] Can both real and test implementations satisfy interfaces cleanly — or do tests require hacks?

### Abstraction Levels
- [ ] Does any interface mix high-level and low-level operations? (confused responsibilities)
- [ ] Do high-level interfaces expose low-level details? (leaky abstractions)
- [ ] Is business logic mixed with infrastructure concerns in one interface?

### Context Usage
- [ ] Do all I/O methods accept `context.Context` as first parameter?
- [ ] Do any non-I/O methods accept context unnecessarily? (API noise)

## Phase 3: Composition Assessment — Where Is State Confused?

### Processors (Should Be Stateless)
- [ ] Identify types that transform data — are they actually stateless?
- [ ] Do any "processors" have internal mutable state? (hidden side effects)
- [ ] Are processor methods deterministic for same input?
- [ ] Can processors be tested without setup/teardown? (if not, they're connectors in disguise)

### Connectors (Should Be Stateful)
- [ ] Identify types that manage state or perform I/O
- [ ] Do connectors own their lifecycle (open/close, connect/disconnect)?
- [ ] Is connector state encapsulated — or leaked via public fields?
- [ ] Do connectors accept dependencies via constructors — or create them internally?

### Values (Should Be Simple)
- [ ] Identify simple data types
- [ ] Do any value types carry hidden state?
- [ ] Do any value types have behavior beyond basic formatting/validation?

### Separation
- [ ] Is there a clear line between processors, connectors, and values — or are there hybrids?
- [ ] Do dependencies flow inward (connectors use processors) — or is it inverted?

## Phase 4: Boundary Assessment — Where Are the Open Doors?

### Boundary Identification
- [ ] External input boundaries identified (user input, HTTP, CLI) — or missing?
- [ ] Storage boundaries identified (database, filesystem, cache) — or implicit?
- [ ] Output boundaries identified (HTTP responses, logs, events) — or leaking?
- [ ] Inter-service boundaries identified (API calls, message queues) — or blurred?

### Boundary Behavior
- [ ] Does each boundary have explicit transformation logic — or is data passed through raw?
- [ ] Do input boundaries validate and sanitize — or trust blindly?
- [ ] Do output boundaries format and redact — or expose internals?
- [ ] Do storage boundaries handle serialization explicitly — or rely on defaults?

### Boundary Clarity
- [ ] Is boundary code identifiable — or buried in business logic?
- [ ] Are boundary transformations testable in isolation — or entangled?
- [ ] Are there implicit conversions happening outside boundaries? (hidden data transformation)

## Phase 5: Dependency Assessment — Where Is Trust Misplaced?

### Production Dependencies
- [ ] Read `go.mod` — how many direct dependencies?
- [ ] Is each dependency justified — or could stdlib handle it?
- [ ] Are there dependencies for trivial functionality? (string helpers, simple math)
- [ ] Are provider-specific SDKs in root module — or properly isolated in submodules?

### Dependency Direction
- [ ] Do dependencies flow inward (concrete depends on abstract)?
- [ ] Does core package depend on providers? (architectural inversion)
- [ ] Are there circular dependencies? (structural fragility)

### Zoobzio Ecosystem
- [ ] Could any external dep be replaced by a zoobzio package?
- [ ] Are patterns consistent with sibling packages?

## Phase 6: Error Assessment — Where Do Failures Hide?

### Sentinel Errors
- [ ] Are expected failure modes covered by package-level sentinel errors — or ad hoc strings?
- [ ] Do sentinel errors use consistent naming (`Err[Condition]`)?
- [ ] Are sentinel errors documented?

### Error Wrapping
- [ ] Are errors wrapped with context using `fmt.Errorf` and `%w` — or swallowed?
- [ ] Is the error chain preserved — or do bare `errors.New` calls discard context?
- [ ] Do error messages describe what failed with actionable detail?

### Error Consistency
- [ ] Do same conditions produce same error types across implementations?
- [ ] Can consumers check errors with `errors.Is`/`errors.As` — or do they need string matching?

## Phase 7: Type Safety Assessment — Where Is the Compiler Bypassed?

### Generics Usage
- [ ] Are type parameters used where values vary by type — or is `interface{}` used instead?
- [ ] Are type constraints as narrow as possible?

### Compile-Time Guarantees
- [ ] Are invalid states unrepresentable where possible?
- [ ] Are type assertions minimized — or sprinkled everywhere?
- [ ] Is interface satisfaction verified at compile time (`var _ Interface = (*Type)(nil)`)?

## Phase 8: Observability Assessment — Where Is the System Blind?

### Identity
- [ ] Do components have names or identifiers — or are they anonymous?
- [ ] Is identity consistent and traceable?

### Signals
- [ ] Can observable events be emitted — or is the system silent?
- [ ] Is there a hard dependency on specific observability infrastructure? (coupling)
- [ ] Is correlation possible across component boundaries?

## Phase 9: Surface Architecture — Cross-Surface Verification

### Shared vs Surface-Specific
- [ ] Stores are shared — same implementation satisfies contracts on both surfaces
- [ ] Contracts are surface-specific — each surface defines its own interface
- [ ] Wire types differ appropriately (api/ masks, admin/ exposes)
- [ ] Handlers are surface-specific — no cross-surface handler leakage
- [ ] Transformers are surface-specific — mapping logic matches surface's wire types
- [ ] Surface-specific artifacts in correct directory (not in wrong surface)

## Phase 10: Entity Completeness — Is Every Chain Complete?

For each domain entity:
- [ ] Model exists in `models/`
- [ ] Migration exists in `migrations/` (if database-backed)
- [ ] Contract exists in `{surface}/contracts/`
- [ ] Store exists in `stores/`
- [ ] Wire types exist in `{surface}/wire/`
- [ ] Transformers exist in `{surface}/transformers/`
- [ ] Handlers exist in `{surface}/handlers/`
- [ ] Each artifact references the correct types from adjacent layers

## Phase 11: Registration Wiring — Is Everything Connected?

### Cross-Reference
- [ ] Every store registered in `stores/stores.go`
- [ ] Every handler registered in `{surface}/handlers/handlers.go`
- [ ] Domain errors mapped in `{surface}/handlers/errors.go`
- [ ] Wire boundaries registered in `{surface}/wire/boundary.go`
- [ ] Model boundaries registered in `models/boundary.go`
- [ ] Boundary registration happens before `sum.Freeze()` (ordering matters)

## Phase 12: Migration Ordering — Do Dependencies Resolve?

- [ ] Migrations ordered by entity dependency
- [ ] Foreign key references point to tables created in earlier migrations
- [ ] No circular migration dependencies

## Phase 13: Cross-Cutting — Where Are the Inconsistencies?

### Consistency
- [ ] Are naming conventions consistent across the package — or do they drift?
- [ ] Are constructor patterns consistent — or mixed?
- [ ] Are error handling patterns consistent — or ad hoc?

### Package Organization
- [ ] Do file names reflect their contents?
- [ ] Is related functionality colocated — or scattered?
- [ ] Any single file doing too many things?
