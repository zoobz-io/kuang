# Testing Create Checklist

## Phase 1: Discovery

### Application Understanding
- [ ] List all `.go` files in shared layers (models/, stores/)
- [ ] List all `.go` files in api/ surface
- [ ] List all `.go` files in admin/ surface
- [ ] Identify which files need corresponding `_test.go` files
- [ ] Review existing tests if any
- [ ] Identify domain-specific testing needs

### Scope Determination
- [ ] Ask: "Is this a new test setup or adding to existing?"
- [ ] Ask: "What external dependencies need integration tests?" (databases, APIs, etc.)
- [ ] Ask: "What operations need benchmarking?"

## Phase 2: Unit Tests

### 1:1 Mapping
For each source file, ensure a corresponding test file exists:

**Shared Layers:**
- [ ] `models/[entity].go` → `models/[entity]_test.go`
- [ ] `stores/[entity].go` → `stores/[entity]_test.go`

**Public API Surface:**
- [ ] `api/handlers/[entity].go` → `api/handlers/[entity]_test.go`
- [ ] `api/transformers/[entity].go` → `api/transformers/[entity]_test.go`

**Admin API Surface:**
- [ ] `admin/handlers/[entity].go` → `admin/handlers/[entity]_test.go`
- [ ] `admin/transformers/[entity].go` → `admin/transformers/[entity]_test.go`

**Exception:** Files containing only delegation, re-exports, or trivial wiring may omit tests if there is no testable logic.

### Test Conventions
- [ ] Tests use `testing` package
- [ ] Table-driven tests for multiple cases
- [ ] Test both success and failure paths
- [ ] Use `t.Run()` for subtests
- [ ] Use `t.Parallel()` where safe

### Test Commands
Tests should run with:
```bash
go test -v -race ./...
```

## Phase 3: testing/ Directory

### Structure
- [ ] Create `testing/` directory
- [ ] Create `testing/README.md`
- [ ] Create `testing/fixtures.go`
- [ ] Create `testing/fixtures_test.go`
- [ ] Create `testing/mocks.go`
- [ ] Create `testing/mocks_test.go`
- [ ] Create `testing/helpers.go`
- [ ] Create `testing/helpers_test.go`
- [ ] Create `testing/benchmarks/` directory
- [ ] Create `testing/benchmarks/README.md`
- [ ] Create `testing/integration/` directory
- [ ] Create `testing/integration/README.md`
- [ ] Create `testing/integration/setup.go`

### testing/README.md

```markdown
# Testing

Overview of testing strategy for [application].

## Running Tests

```bash
make test              # All tests with race detector
make test-unit         # Unit tests only (short mode)
make test-integration  # Integration tests
make test-bench        # Benchmarks
```

## Structure

- `fixtures.go` — Test data factories
- `mocks.go` — Contract mock implementations
- `helpers.go` — Domain-specific test utilities
- `benchmarks/` — Performance tests
- `integration/` — End-to-end tests

## Coverage

Target: 70% project, 80% new code
```

## Phase 4: Fixtures

### fixtures.go

```go
package testing

import (
    "testing"
    "github.com/zoobzio/[app]/models"
)

type UserOption func(*models.User)

func NewUser(t *testing.T, opts ...UserOption) *models.User {
    t.Helper()
    u := &models.User{
        ID:    "test-user-id",
        Name:  "Test User",
        Email: "test@example.com",
    }
    for _, opt := range opts {
        opt(u)
    }
    return u
}

func WithUserID(id string) UserOption {
    return func(u *models.User) {
        u.ID = id
    }
}
```

### Fixture Conventions
- [ ] Factory functions return domain models
- [ ] Sensible defaults provided
- [ ] Option pattern for customization
- [ ] Factory functions call `t.Helper()`

### fixtures_test.go
- [ ] Every fixture function has tests
- [ ] Options are tested

## Phase 5: Mocks

### mocks.go

```go
package testing

import (
    "context"
    "github.com/zoobzio/[app]/models"
)

type MockUsers struct {
    OnGet    func(ctx context.Context, id string) (*models.User, error)
    OnCreate func(ctx context.Context, u *models.User) error
    OnList   func(ctx context.Context) ([]*models.User, error)
}

func (m *MockUsers) Get(ctx context.Context, id string) (*models.User, error) {
    if m.OnGet != nil {
        return m.OnGet(ctx, id)
    }
    return nil, nil
}
```

### Mock Conventions
- [ ] Function-field pattern
- [ ] Implements contract interfaces
- [ ] Default behavior is nil/zero values
- [ ] Each contract has a mock

### mocks_test.go
- [ ] Mocks implement interfaces (compile-time check)
- [ ] Mock behavior is testable

## Phase 6: Helpers

### helpers.go

```go
package testing

import "testing"

func AssertNoError(t *testing.T, err error) {
    t.Helper()
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}

func AssertError(t *testing.T, err error) {
    t.Helper()
    if err == nil {
        t.Fatal("expected error, got nil")
    }
}
```

### Helper Conventions
- [ ] All helpers call `t.Helper()` first
- [ ] All helpers accept `*testing.T` as first parameter
- [ ] Helpers are domain-specific, not generic
- [ ] Assertions follow `Assert[Condition]` naming
- [ ] Setup functions follow `New[Thing]` or `With[Thing]` naming

### helpers_test.go
- [ ] Every helper function has tests
- [ ] Test both passing and failing cases

## Phase 7: Benchmarks

### testing/benchmarks/README.md

```markdown
# Benchmarks

Performance tests for [application].

## Running

```bash
make test-bench
# or
go test -bench=. -benchmem ./testing/benchmarks/...
```

## Results

Document baseline results here for comparison.
```

### Benchmark Conventions
- [ ] File naming: `[feature]_test.go`
- [ ] Function naming: `Benchmark[Operation]`
- [ ] Include memory benchmarks (`-benchmem`)
- [ ] Document baseline results in README

## Phase 8: Integration Tests

### testing/integration/README.md

```markdown
# Integration Tests

End-to-end tests for [application].

## Prerequisites

- Database running (see docker-compose.yml)
- Environment configured (see .env.example)

## Running

```bash
make test-integration
# or
go test -v -race ./testing/integration/...
```

## Test Isolation

Each test should set up and tear down its own state.
```

### testing/integration/setup.go

```go
package integration

import (
    "testing"
    "github.com/zoobzio/[app]/stores"
)

type TestRegistry struct {
    stores *stores.Stores
}

type Option func(*TestRegistry)

func NewTestRegistry(t *testing.T, opts ...Option) *TestRegistry {
    t.Helper()
    r := &TestRegistry{}
    for _, opt := range opts {
        opt(r)
    }
    return r
}

func WithUsers(users stores.Users) Option {
    return func(r *TestRegistry) {
        // configure users store
    }
}
```

### Integration Test Conventions
- [ ] File naming: `[feature]_test.go`
- [ ] Tests are self-contained (setup/teardown)
- [ ] Skip if dependencies unavailable: `t.Skip("requires [dependency]")`
- [ ] Use build tags if needed: `//go:build integration`

## Phase 9: CI Coverage Configuration

Integration tests must contribute to coverage reports.

### Coverage Workflow Updates

Add to `.github/workflows/coverage.yml`:

```yaml
- name: Run tests with coverage
  run: |
    # Unit test coverage
    go test -v -race -coverprofile=coverage-unit.out -covermode=atomic ./...

    # Integration test coverage (separate profile)
    go test -v -race -coverprofile=coverage-integration.out -covermode=atomic ./testing/integration/...

    # Merge coverage profiles
    echo "mode: atomic" > coverage.out
    tail -n +2 coverage-unit.out >> coverage.out
    tail -n +2 coverage-integration.out >> coverage.out 2>/dev/null || true

- name: Upload coverage
  uses: codecov/codecov-action@v4
  with:
    files: ./coverage.out
    flags: unit,integration
```

### Makefile Updates

Ensure Makefile supports combined coverage:

```makefile
coverage: ## Generate coverage report (unit + integration)
	@go test -coverprofile=coverage-unit.out -covermode=atomic ./...
	@go test -coverprofile=coverage-integration.out -covermode=atomic ./testing/integration/... 2>/dev/null || true
	@echo "mode: atomic" > coverage.out
	@tail -n +2 coverage-unit.out >> coverage.out
	@tail -n +2 coverage-integration.out >> coverage.out 2>/dev/null || true
	@go tool cover -html=coverage.out -o coverage.html
	@go tool cover -func=coverage.out | tail -1
	@echo "Coverage report: coverage.html"
```

## Phase 10: Verification

- [ ] `make test` passes
- [ ] `make test-integration` passes (or skips gracefully)
- [ ] `make test-bench` runs benchmarks
- [ ] `make coverage` generates combined report
- [ ] All source files have corresponding test files
- [ ] Fixtures, mocks, and helpers are tested
- [ ] CI workflow captures integration coverage
