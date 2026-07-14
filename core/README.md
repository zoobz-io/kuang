# core

The kuang framework entrypoint and module-facing endpoint helpers.

## Purpose

`core` is what a consumer imports to stand up a kuang server. It owns the bootstrap
sequence (`Run`), the `Module` contract that tool packs implement, and the
auth-aware endpoint builders that resolve per-agent upstream credentials.

Nothing in `core` is agent-facing — it wires together `api/` (the built-in
credential API), `internal/auth` (mTLS + identity), and any composed modules.

## Run

```go
func Run(modules ...Module) error
```

`Run` bootstraps the server:

1. Loads `config.Security` and `config.Server` via `fig`.
2. Builds the sctx authority and mTLS `*tls.Config` from the configured certs
   (`internal/auth`).
3. Opens the SQLite credential store (`api/stores`) and registers it as the default
   `contracts.Credentials` in the `sum` registry.
4. Seeds the built-in `/creds` endpoints, then calls each `Module` to collect more.
5. Freezes the registry, wraps the engine with the scoped-OpenAPI interceptor and
   the mTLS-terminating middleware, and serves HTTPS with graceful shutdown on
   SIGINT/SIGTERM.

A consumer's `main` typically does nothing but call it:

```go
core.Run(github.Module(ghCfg), matrix.Module(mxCfg))
```

## The Module contract

```go
type Module func(ctx context.Context, k sum.Key) ([]rocco.Endpoint, error)
```

A module's closure has two jobs, both at registration time:

- **register services** into the `sum.Key` registry (`sum.Register[contracts.API](k, svc)`)
- **return its rocco endpoints**

The registry is frozen after every module runs, so all registration must happen
inside the returned closure. Modules discover the credential store through `sum`, not
through parameters — see [Per-Agent Credentials](#credentials) below.

## Authenticated endpoint helpers

Modules build endpoints with `core.GET/POST/PUT/DELETE` rather than the bare `rocco`
equivalents. These wrap rocco and resolve the calling agent's upstream token before
the handler runs:

```go
func GET[In, Out any](
    path, credentialKey string,
    fn func(*rocco.Request[In], httpc.RequestOption) (Out, error),
) *rocco.Handler[In, Out]
```

- `credentialKey` names which stored secret to resolve (e.g. `"github-token"`).
- Before `fn` runs, `ResolveBearer` looks up the token for the calling agent's
  identity and passes it as an `httpc.RequestOption` (an `Authorization: Bearer`
  header) — always the last argument to `fn`.
- The returned `*rocco.Handler` keeps the fluent builder, so handlers chain
  `.WithName().WithSummary().WithTags().WithPathParams().WithScopes(...)`.

<a name="credentials"></a>

## ResolveBearer

```go
func ResolveBearer(ctx context.Context, credentialKey string) (httpc.RequestOption, error)
```

The bridge from **mTLS identity** to **upstream credential**: it pulls the
`contracts.Credentials` store from `sum`, reads the caller via
`auth.IdentityFromContext(ctx)`, resolves `Resolve(ctx, identity.ID(), credentialKey)`,
and returns `httpc.WithRequestBearerToken(token)`. Failures short-circuit as
`rocco.ErrForbidden` ("credentials not configured") or `ErrUnauthorized`.

This means each agent acts under **its own** upstream identity: the operator seeds a
per-agent token via the `/creds` endpoints, and every module call for that agent
carries that agent's token.

## Files

| File | Contents |
|------|----------|
| `kuang.go` | `Run`, the `Module` type, the bootstrap sequence |
| `endpoints.go` | `ResolveBearer` + the `GET/POST/PUT/DELETE` helpers |
