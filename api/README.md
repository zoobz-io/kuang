# api

kuang's built-in **per-agent credential API** — the one domain the framework ships
itself, sliced one file per architectural layer.

## Purpose

`api/` is the reference implementation of a kuang domain and, at the same time, a
working feature: the `/creds` endpoints let an agent manage the upstream tokens it
owns, and let modules resolve those tokens at request time. Modules follow the same
layering (see [`modules/README.md`](../modules/README.md)).

Every operation is scoped by `agent`, and the agent is **always** the authenticated
mTLS identity (`identity.ID()`), never a value from the request body or query. That
is what makes the store a per-agent isolation boundary.

## Layers

The dependency direction points inward only — nothing in `api/` imports `core/` or
`modules/`.

| Package     | Role                                                             | Key symbols                                                                             |
| ----------- | ---------------------------------------------------------------- | --------------------------------------------------------------------------------------- |
| `config`    | Typed config loaded via `fig`, validated with `check`            | `Security`, `Server`                                                                    |
| `contracts` | The interface handlers depend on (DI seam; modules may override) | `Credentials`                                                                           |
| `models`    | `grub` persistence row                                           | `Credential`                                                                            |
| `wire`      | JSON request/response DTOs, distinct from the DB model           | `CredentialResponse`, `SetCredentialRequest`, `CredentialKeysResponse`, `CredentialRef` |
| `stores`    | Concrete SQLite `Credentials` (sqlx + grub + astql)              | `NewCredentials`, `*Credentials`                                                        |
| `handlers`  | The `/creds` rocco endpoints                                     | `GetCredential`, `ListCredentials`, `SetCredential`, `DeleteCredential`                 |
| `events`    | capitan signal vocabulary (credential store + startup)           | `CredentialResolved`, `StartupServerListening`, …                                       |
| `extend`    | rocco engine override serving the scope-filtered OpenAPI spec    | `OpenAPIInterceptor`                                                                    |

## The credential contract

```go
type Credentials interface {
    Resolve(ctx context.Context, agent, key string) (string, error)
    Set(ctx context.Context, agent, key, value string) error
    Delete(ctx context.Context, agent, key string) error
    List(ctx context.Context, agent string) ([]string, error)
}
```

Two consumer roles:

- **Modules** call `Resolve` to get an upstream token (GitHub PAT, Matrix access
  token) for the calling agent — via `core.ResolveBearer`.
- **Agents** call the `/creds` endpoints (`handlers`) to manage the credentials they
  own.

`core.Run` registers `stores.Credentials` as the default implementation. A module can
override it by registering its own `contracts.Credentials` — `sum` replaces silently
on duplicate registration. (kuang never picks a store for you beyond this default;
composition is the operator's job.)

## Endpoints

| Method + path            | Handler            | Scope          |
| ------------------------ | ------------------ | -------------- |
| `GET /creds/{key...}`    | `GetCredential`    | `creds-read`   |
| `GET /creds`             | `ListCredentials`  | `creds-read`   |
| `PUT /creds/{key...}`    | `SetCredential`    | `creds-write`  |
| `DELETE /creds/{key...}` | `DeleteCredential` | `creds-delete` |

`{key...}` is a rocco wildcard so keys may contain slashes. Read/write/delete are
separate scopes so a read-only grant is possible. `List` returns keys only — values
never leave the store through it.

## The scoped-OpenAPI interceptor

```go
func OpenAPIInterceptor(engine *rocco.Engine) http.Handler
```

`extend` wraps the rocco router (rather than registering a `GET /openapi` endpoint,
which rocco reserves). On `GET /openapi` it reads the caller's identity and returns
`engine.GenerateOpenAPI(identity)` — a spec filtered to the endpoints the agent's
scopes allow. This is the core of the "agents only discover tools they're authorized
for" story. `core.Run` installs it as the inner handler, wrapped by
`auth.Terminate`.

## Notes

- `api/config/README.md` documents the config conventions in detail.
- The `events` **credential-store** signals (`creds.*`) are emitted by `api/stores` on
  every Resolve/Set/Delete/List (and on not-found/failure). The secret value is never
  included in a signal — only agent, key name, count, operation, and error.
- The `events` **startup** signals (`startup.*`) are declared but **not yet emitted** —
  a forward-looking vocabulary; `core.Run` currently logs startup with the standard
  library.
