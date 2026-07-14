# kuang

An mTLS security boundary that exposes **scope-filtered tools to AI agents** over
[MCP](https://modelcontextprotocol.io).

## Overview

kuang sits between AI agents and the upstream APIs they need (GitHub, Matrix, …).
Each agent runs in its own container, holds its own client certificate, and talks
to kuang over mutual TLS. The certificate **is** the agent's identity: its
CommonName names the agent and its Organizational Unit entries carry the scopes it
has been granted. kuang terminates the mTLS connection, resolves that identity, and
serves the agent an OpenAPI spec filtered to exactly the operations its scopes
allow. An MCP bridge turns that spec into MCP tools, so an agent only ever discovers
and calls tools it is authorized for.

kuang is a **framework, not a deployable app**. You write a small Go `main` that
imports kuang and composes the modules you want:

```go
package main

import (
    "log"

    "github.com/zoobzio/kuang/core"
    "github.com/zoobzio/kuang/modules/github"
    "github.com/zoobzio/kuang/modules/matrix"
    githubcfg "github.com/zoobzio/kuang/modules/github/config"
    matrixcfg "github.com/zoobzio/kuang/modules/matrix/config"
)

func main() {
    err := core.Run(
        github.Module(githubcfg.API{Owner: "zoobzio"}),
        matrix.Module(matrixcfg.API{Homeserver: "https://matrix.org"}),
    )
    if err != nil {
        log.Fatal(err)
    }
}
```

Install only the modules you need — each is a separate Go module.

## How it fits together

```
┌─────────────┐   stdio    ┌────────────┐    mTLS     ┌─────────────────────┐
│  AI agent   │◀──MCP────▶ │  cmd/mcp   │◀──HTTPS───▶ │   kuang (core.Run)  │
│ (container) │            │  (bridge)  │             │                     │
└─────────────┘            └────────────┘             │  auth.Terminate     │  cert → identity + scopes
      │ owns cert                                     │  extend.OpenAPI     │  spec filtered by scope
      │ (CN=agent, OU=scopes)                         │  modules → handlers │
                                                      │  ResolveBearer      │  per-agent upstream token
                                                      └──────────┬──────────┘
                                                                 │ httpc (Bearer <agent token>)
                                                                 ▼
                                                        upstream APIs (GitHub, Matrix, …)
```

1. The MCP bridge (`cmd/mcp`) fetches `GET /openapi` over mTLS. Because the request
   carries the agent's cert, kuang returns a **scope-filtered** spec.
2. Each OpenAPI operation becomes one MCP tool. `tools/call` is proxied back to
   kuang as the corresponding HTTP request.
3. A module handler resolves the calling agent's stored upstream token
   (keyed by identity + credential key) and calls the third-party API on its behalf.

See the memory notes and package READMEs for the full design rationale.

## Project structure

```
kuang/
├── core/              # Framework entrypoint: core.Run + auth-aware endpoint helpers
├── api/               # Built-in per-agent credential API
│   ├── config/        #   Security + Server config (fig)
│   ├── contracts/     #   Credentials interface (DI seam)
│   ├── models/        #   grub persistence model
│   ├── wire/          #   HTTP request/response DTOs
│   ├── stores/        #   SQLite credential store
│   ├── handlers/      #   /creds endpoints
│   ├── events/        #   capitan startup signals
│   └── extend/        #   Scoped-OpenAPI interceptor
├── internal/
│   ├── auth/          # mTLS termination + sctx identity/scopes
│   ├── httpc/         # Instrumented outbound HTTP client
│   ├── mcp/           # OpenAPI → MCP bridge (stdio)
│   └── otel/          # OpenTelemetry providers
├── modules/           # Pluggable tool packs (separate Go modules)
│   ├── github/        #   GitHub REST tools
│   └── matrix/        #   Matrix client-server tools
└── cmd/mcp/           # The MCP bridge binary
```

Each directory carries a README explaining its role.

## Configuration

kuang is configured entirely through environment variables (loaded via `fig`).

**Server** (`api/config`):

| Variable | Default | Purpose |
|----------|---------|---------|
| `KUANG_HOST` | `localhost` | Listen host |
| `KUANG_PORT` | `8080` | Listen port |
| `KUANG_DB_PATH` | `data/kuang.db` | SQLite credential store (auto-created) |

**Security / mTLS** (`api/config`):

| Variable | Default | Purpose |
|----------|---------|---------|
| `APP_CA_CERT_PATH` | `certs/ca.pem` | CA used to verify client (agent) certs |
| `APP_CERT_PATH` | `certs/server.pem` | Server leaf certificate |
| `APP_KEY_PATH` | `certs/server-key.pem` | Server private key |
| `APP_CRYPTO_ALGO` | `ed25519` | `ed25519` or `ecdsa-p256` |
| `APP_REQUIRE_MTLS` | `true` | `true` → require & verify client cert |

**MCP bridge** (`cmd/mcp`): `KUANG_URL`, `KUANG_CA_CERT`, `KUANG_CERT`, `KUANG_KEY`
— see [`cmd/mcp/README.md`](cmd/mcp/README.md).

**Telemetry** (`internal/otel`): `OTEL_EXPORTER_OTLP_ENDPOINT`, `OTEL_SERVICE_NAME`.

kuang does **not** issue certificates. It trusts a CA and expects server/agent certs
to be provisioned externally (e.g. by a step-ca provisioner that stamps agent scopes
into the cert OU). Point `APP_CA_CERT_PATH` at that CA.

## Development

### Prerequisites

- Go 1.25+
- golangci-lint v2.7.2

```bash
make install-tools    # golangci-lint
make build            # build the MCP bridge (bin/kuang-mcp)
make test             # go test -race ./...
make lint             # golangci-lint
```

`internal/otel` can export traces/logs/metrics to any OTLP endpoint via
`OTEL_EXPORTER_OTLP_ENDPOINT` — point it at your own collector.

> **Known gap:** the build/run tooling (`make build`, `make run`, `.air.toml`,
> `.goreleaser.yml`, and the `app`/`postgres`/`redis`/`minio`/`migrate` services in
> `docker-compose.yml`) still references the pre-refactor template layout
> (`cmd/app`, `admin/`, `testing/`, `migrations/`) and is **not wired to the current
> code**. kuang no longer ships a server `main` — `core.Run` is a library entrypoint
> that a consumer's `main` calls (see the Overview). These files should be
> reconciled with the real layout in a follow-up.

## License

MIT
