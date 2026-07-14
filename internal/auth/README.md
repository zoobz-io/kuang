# internal/auth

mTLS termination and sctx-based authorization: turn a client certificate into an
agent identity carrying scopes.

## Purpose

This package is the security core. It bootstraps an [sctx](https://github.com/zoobz-io/sctx)
authority from the server's certificates, terminates each agent's mTLS connection,
and exposes the resulting identity (name + scopes) to the rest of the server.

## The cert → identity → scope mapping

An agent's client certificate carries its authorization:

- **Subject CommonName (CN)** → the agent's identity / name (`Identity.ID()`).
- **Subject Organizational Unit (OU) entries** → the granted permission scopes
  (`Identity.Scopes()`).

For example a cert with `OU=github-repos-read, OU=creds-read` grants exactly those two
scopes. This mapping lives in the (unexported) `agentPolicy`, which seeds
`base.Permissions = cert.Subject.OrganizationalUnit`. kuang does not issue certs — a
CA (e.g. a step-ca provisioner) stamps these OUs based on the agent name.

## Bootstrap — `admin.go`

```go
func New(ctx context.Context, cfg config.Security) (*Authority, error)
func TLSConfig(cfg config.Security) (*tls.Config, error)
```

- `New` reads the CA into a pool, loads the server keypair, and builds an
  `sctx.NewAdminService[AgentMeta]` with the agent policy and a 256-entry bounded
  cache. It mints the server's own trusted token — the permission ceiling for guard
  creation.
- `TLSConfig` returns a TLS 1.3-minimum config with the CA as `ClientCAs`;
  `ClientAuth` is `RequireAndVerifyClientCert` when `RequireMTLS`, else
  `VerifyClientCertIfGiven`.

`type Authority struct { Admin sctx.Admin[AgentMeta]; Token sctx.SignedToken }`.
`type AgentMeta struct { Agent string }` rides along in every security context.

## Middleware — `middleware.go`

```go
func Terminate(authority *Authority) func(http.Handler) http.Handler
func IdentityFromContext(ctx context.Context) *Identity
func Authenticator() func(context.Context, *http.Request) (rocco.Identity, error)
```

- `Terminate` is the outermost middleware. It rejects requests with no client cert
  (`ErrUnauthorized`), calls `admin.GenerateTrusted` (the handshake already proved key
  possession), retrieves the security context by cert fingerprint, and injects an
  `*Identity` into the request context. `core.Run` wraps the whole handler chain with
  it.
- `IdentityFromContext` is the primary API for handlers — returns the injected
  `*Identity` (or nil). Used by `api/handlers`, `api/extend`, and `core.ResolveBearer`.
- `Authenticator` adapts the injected identity into a rocco identity-extractor;
  registered via `engine.WithAuthenticator`. rocco enforces each endpoint's
  `.WithScopes(...)` against `Identity.HasScope`.

`*Identity` implements `rocco.Identity`: `ID()`→CN, `TenantID()`→agent name,
`Scopes()`→OU permissions, `HasScope()`→`ctx.HasPermission`.

## Admin user store — `store.go`

An SQLite-backed store (`grub` + bcrypt) with `users` and `meta` tables:
`OpenStore`, `Create`, `Verify` (constant-time on missing user), `Initialized`,
`Initialize` (transactional first-run setup).

> **Not yet wired.** `store.go`, `ServerPrivateKey`, and `Authority.Token` are defined
> but not referenced outside this package — scaffolding for a future admin-user API
> and assertion verification, not part of the current `core.Run` path.
