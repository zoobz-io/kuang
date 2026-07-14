# config

Typed configuration structs, loaded from the environment via `fig` and validated
with `check`.

## Purpose

Define flat, single-purpose config structs populated from environment variables. Each
struct implements `Validate() error` so bad configuration fails fast at boot. `core`
loads these with `fig.Load`.

## Structs

### `Security` (`security.go`) — TLS / mTLS

| Field | Env | Default | Notes |
|-------|-----|---------|-------|
| `CACertPath` | `APP_CA_CERT_PATH` | `certs/ca.pem` | CA that signs agent (client) certs |
| `CertPath` | `APP_CERT_PATH` | `certs/server.pem` | Server leaf cert |
| `KeyPath` | `APP_KEY_PATH` | `certs/server-key.pem` | Server private key |
| `CryptoAlgo` | `APP_CRYPTO_ALGO` | `ed25519` | `ed25519` or `ecdsa-p256` |
| `RequireMTLS` | `APP_REQUIRE_MTLS` | `true` | `true` → require & verify client cert |

`Validate()` requires the three paths and constrains `CryptoAlgo` to the two allowed
values.

### `Server` (`server.go`)

| Field | Env | Default | Notes |
|-------|-----|---------|-------|
| `Host` | `KUANG_HOST` | `localhost` | Listen host |
| `DBPath` | `KUANG_DB_PATH` | `data/kuang.db` | SQLite credential store (auto-created) |
| `Port` | `KUANG_PORT` | `8080` | Listen port |

`Validate()` requires host and db path and checks the port range.

## Pattern

```go
type Server struct {
    Host   string `env:"KUANG_HOST" default:"localhost"`
    DBPath string `env:"KUANG_DB_PATH" default:"data/kuang.db"`
    Port   int    `env:"KUANG_PORT" default:"8080"`
}

func (c Server) Validate() error {
    return check.All(
        check.Str(c.Host, "host").Required().V(),
        check.Str(c.DBPath, "db_path").Required().V(),
        check.Int(c.Port, "port").PortNumber().V(),
    ).Err()
}
```

Loaded in `core.Run`:

```go
var srvCfg config.Server
if err := fig.Load(&srvCfg); err != nil {
    return fmt.Errorf("load server config: %w", err)
}
```

## Struct tags

| Tag | Purpose | Example |
|-----|---------|---------|
| `env` | Environment variable name | `env:"KUANG_PORT"` |
| `default` | Default if unset | `default:"8080"` |
| `secret` | Secret path for a provider (e.g. Vault) | `secret:"app/db-password"` |

## Guidelines

- One struct per config domain; prefer separate flat structs over nesting.
- Always implement `Validate() error` using `check`.
- Use the `secret` tag for credentials, keys, and tokens.
- Add helper methods (e.g. an address builder) where they reduce duplication.
