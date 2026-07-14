# internal/httpc

An instrumented outbound HTTP client for calling upstream APIs on behalf of an agent.

## Purpose

This is the client kuang **modules** use to reach third-party APIs (GitHub, Matrix,
…). Every request emits [capitan](https://github.com/zoobz-io/capitan) signals for
observability. It is distinct from `internal/mcp.Client`, which is the mTLS client the
MCP bridge uses to call the kuang server itself.

## Client

```go
func New(opts ...Option) *Client
```

Client-level options (applied as defaults to every request):

- `WithBaseURL(url)`
- `WithHeader(key, value)`
- `WithBearerToken(token)`
- `WithTimeout(d)` — default 30s

Verb helpers all take `(ctx, path, [body,] opts ...RequestOption)` and return
`(*Response, error)`: `Get`, `Post`, `Put`, `Patch`, `Delete`.

```go
client := httpc.New(
    httpc.WithBaseURL("https://api.github.com"),
    httpc.WithHeader("Accept", "application/vnd.github+json"),
)
resp, err := client.Get(ctx, "/repos/zoobzio/kuang", auth)
```

`Response` carries `Status`, `Headers`, `Body`, and `DurationMs`, with
`Decode(v any) error` for JSON.

## Per-request options — the credential mechanism

```go
type RequestOption func(req *http.Request)

func WithRequestHeader(key, value string) RequestOption
func WithRequestBearerToken(token string) RequestOption
```

Per-request options are applied **after** client defaults, so they win. This is how
each agent call carries that agent's own upstream token:
`core.ResolveBearer` returns a `WithRequestBearerToken(...)` for the calling agent,
and module handlers thread it through to the service call.

## Observability — `events.go`

Signals: `RequestStarted`, `RequestCompleted`, `RequestError` (4xx/5xx),
`RequestFailed` (transport error). Field keys: `MethodKey`, `URLKey`, `StatusKey`,
`DurationMsKey`, `ErrorKey`. A status ≥ 400 returns both the `*Response` and an error
whose message includes the body.
