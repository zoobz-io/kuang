# Testing Create

Set up test infrastructure for a Go application: unit tests, helpers, mocks, fixtures, benchmarks, integration tests, and CI coverage configuration.

## Principles

1. **1:1 mapping** — Every source file MUST have a corresponding test file
2. **Colocated tests** — Unit tests live next to source, infrastructure lives in `testing/`
3. **Surface-aware** — Tests respect API surface boundaries
4. **Coverage includes integration** — CI MUST capture coverage from integration tests

## Execution

1. Read `checklist.md` in this skill directory
2. Determine scope: full setup or adding to existing?
3. Ask user about application-specific testing needs
4. Create files per specifications
5. Verify with `make test`

## Specifications

### Directory Structure

```
/
├── models/
│   ├── user.go
│   └── user_test.go
├── stores/
│   ├── users.go
│   └── users_test.go
├── api/
│   ├── handlers/
│   │   ├── users.go
│   │   └── users_test.go
│   └── transformers/
│       ├── users.go
│       └── users_test.go
├── admin/
│   ├── handlers/
│   │   ├── users.go
│   │   └── users_test.go
│   └── transformers/
│       ├── users.go
│       └── users_test.go
└── testing/
    ├── README.md
    ├── fixtures.go
    ├── fixtures_test.go
    ├── mocks.go
    ├── mocks_test.go
    ├── helpers.go
    ├── helpers_test.go
    ├── benchmarks/
    │   ├── README.md
    │   └── core_test.go
    └── integration/
        ├── README.md
        ├── setup.go
        └── [feature]_test.go
```

### 1:1 Mapping Rule

Every `.go` file MUST have a corresponding `_test.go` file.

Exception: Files containing only delegation, re-exports, or trivial wiring with no testable logic.

### Fixture Requirements

`testing/fixtures.go` MUST:
- Return test data with sensible defaults
- Use option pattern for customization
- Be domain-specific

```go
func NewUser(t *testing.T, opts ...UserOption) *models.User
```

### Mock Requirements

`testing/mocks.go` MUST:
- Use function-field pattern
- Implement contracts

```go
type MockUsers struct {
    OnGet func(ctx context.Context, id string) (*models.User, error)
}
```

### Helper Requirements

`testing/helpers.go` MUST:
- Have all helpers call `t.Helper()` as first statement
- Have all helpers accept `*testing.T` as first parameter
- Be domain-specific, not generic utilities

Naming conventions:
- Assertions: `Assert[Condition]`
- Setup: `New[Thing]` or `With[Thing]`

### Integration Setup

`testing/integration/setup.go` MUST:
- Provide real registry with real stores
- Use option pattern: `WithUsers()`, `WithPosts()`
- Handle setup and teardown

### Coverage Targets

| Metric | Target | Threshold |
|--------|--------|-----------|
| Project | 70% | 1% drop allowed |
| Patch | 80% | 0% drop allowed |

### CI Integration Test Coverage

CI MUST capture integration test coverage. See checklist for exact workflow configuration.

## Prohibitions

DO NOT:
- Create test files without corresponding source files
- Create helpers that don't call `t.Helper()`
- Use third-party test frameworks (use standard `testing` package)
- Skip the `testing/fixtures_test.go`, `testing/mocks_test.go`, or `testing/helpers_test.go` files

## Output

A complete test infrastructure that:
- Enforces 1:1 source-to-test mapping
- Provides domain-specific fixtures, mocks, and helpers
- Includes benchmark and integration test scaffolding
- Captures all test coverage in CI reports
