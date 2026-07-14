# modules

Pluggable tool packs. Each module wraps one upstream API and exposes it to agents as
a set of scoped tools.

## Purpose

A module is how a capability enters kuang. It is a **separate Go module** (its own
`go.mod`) so packs can be developed and versioned independently, and a consumer's
`main` installs only the ones it wants:

```go
core.Run(github.Module(ghCfg), matrix.Module(mxCfg))
```

Shipped modules: [`github`](github/README.md) (GitHub REST) and
[`matrix`](matrix/README.md) (Matrix client-server).

## The module contract

Every module exports one function returning a `core.Module`:

```go
func Module(cfg config.API) core.Module {
    return func(_ context.Context, k sum.Key) ([]rocco.Endpoint, error) {
        if err := cfg.Validate(); err != nil {
            return nil, err
        }
        sum.Register[contracts.API](k, services.NewAPI(cfg))
        return handlers.Endpoints(), nil
    }
}
```

The closure validates config (fail-fast at boot), registers its service under a
`contracts.API` into the `sum` registry, and returns its endpoints. The registry is
frozen after all modules run, so **all registration happens inside the closure**. The
`ctx` argument is unused — registration needs only `k`.

## Anatomy

Every module has the same fixed shape, one file per layer:

| Path | Role |
|------|------|
| `module.go` | Assembly point implementing `core.Module`. |
| `config/api.go` | `API` config struct + `Validate()` — the operator's knobs (base URL, owner/homeserver). Validated at registration. |
| `contracts/api.go` | The `API` interface handlers depend on (dependency inversion; enables mock services). The concrete impl is registered under it. |
| `models/*.go` | JSON request/response DTOs, one file per resource, each with `Validate()` (via `check`). The OpenAPI schema surface. |
| `services/api.go` | Concrete `contracts.API`. Owns an `internal/httpc.Client`, maps interface methods to upstream REST calls, decodes into models. All real logic lives here. |
| `handlers/handlers.go` | `Endpoints() []rocco.Endpoint` — the manifest listing every endpoint var. |
| `handlers/auth.go` | `const CredentialKey = "<module>-token"`. |
| `handlers/scopes.go` | Scope constants + `Scopes() []string`. |
| `handlers/<tool>.go` | One thin endpoint var per tool. |

## How a handler works

```go
var ListRepos = core.GET[rocco.NoBody, models.RepoList](
    "/github/repos", CredentialKey,
    func(r *rocco.Request[rocco.NoBody], auth httpc.RequestOption) (models.RepoList, error) {
        return sum.MustUse[contracts.API](r).ListRepos(r, auth)
    },
).WithName("listRepos").WithSummary("List repositories").WithTags("github").WithScopes(ScopeReposRead)
```

- Built with `core.GET/POST/PUT/DELETE` (not bare `rocco`) so the calling agent's
  token for `CredentialKey` is resolved and passed in as the `auth httpc.RequestOption`
  (always last).
- `sum.MustUse[contracts.API](r)` resolves the service from the request context.
- Path params via `r.Params.Path["x"]`, query via `r.Params.Query["x"]`; declare them
  with `.WithPathParams(...)`/`.WithQueryParams(...)`. Parse numerics with
  `strconv.Atoi`, returning `rocco.ErrBadRequest` on failure.
- Exactly **one scope per endpoint** via `.WithScopes(...)`; `.WithTags(...)` groups
  it in the OpenAPI spec. Use `rocco.NoBody` for empty in/out.

## Scopes and credentials

- **Scopes** follow `<module>-<resource>-<read|write>` (e.g. `github-issues-write`).
  `handlers.Scopes()` aggregates them to set the server's permission ceiling. rocco
  enforces each endpoint's scope against the agent's cert-derived scopes, and the
  scoped-OpenAPI interceptor hides tools the agent can't call.
- **Credentials**: each module hard-codes one `CredentialKey`. At request time `core`
  resolves the calling agent's stored token for that key and injects it as a bearer
  token. The operator must seed each agent's `<module>-token` via the `/creds`
  endpoints, or calls 403 with "credentials not configured".

## Writing a new module

1. `modules/<name>/go.mod` — `module github.com/zoobzio/kuang/modules/<name>`, require
   `check`/`kuang`/`rocco`/`sum`, `replace github.com/zoobzio/kuang => ../../`. Add the
   dir to the root `go.work`.
2. `config/api.go` — `API` struct + `Validate()`.
3. `contracts/api.go` — `API` interface; every method takes `ctx` +
   `...httpc.RequestOption`, returns a `models` type.
4. `services/api.go` — `NewAPI(cfg) *API` building `httpc.New(WithBaseURL(...), ...)`;
   implement each method with `client.Get/Post/Put` + `resp.Decode`.
5. `models/*.go` — request/response types with JSON tags and `Validate()`.
6. `handlers/auth.go` — `const CredentialKey = "<name>-token"`.
7. `handlers/scopes.go` — scope constants + `Scopes()`.
8. `handlers/<tool>.go` — one endpoint var each.
9. `handlers/handlers.go` — `Endpoints()` listing every var.
10. `module.go` — the `Module(cfg)` assembly function.

**Gotchas:** services must never assume a token — it always arrives per-request via
the `auth` option; one scope per endpoint; do all `sum.Register` inside the closure.
