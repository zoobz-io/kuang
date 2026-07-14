# cmd/mcp

The kuang MCP bridge — a stdio MCP server that proxies to a running kuang server over
mTLS.

## Purpose

This is the binary an AI agent's MCP host launches. It speaks
[MCP](https://modelcontextprotocol.io) as JSON-RPC over stdin/stdout, and on the other
side is an mTLS client to a kuang server. It is a **client**, not the kuang server —
it does not call `core.Run` or register modules.

On startup it:

1. Builds an mTLS client from the configured cert paths.
2. Fetches kuang's `GET /openapi` — because the request carries the agent's client
   cert, kuang returns a spec **filtered to that agent's scopes**.
3. Turns each operation into an MCP tool (logs `kuang-mcp: loaded N tools`).
4. Serves the JSON-RPC loop on stdio.

So the agent only ever sees and calls tools its certificate authorizes. The bridge is
generic (see [`internal/mcp`](../../internal/mcp/README.md)) — it reflects whatever
kuang exposes.

## Configuration

All via environment variables (no flags):

| Variable | Default | Purpose |
|----------|---------|---------|
| `KUANG_URL` | `https://localhost:8080` | kuang server base URL |
| `KUANG_CA_CERT` | `certs/ca.pem` | CA that signed the kuang server cert |
| `KUANG_CERT` | `certs/client.pem` | This agent's client certificate |
| `KUANG_KEY` | `certs/client-key.pem` | This agent's private key |

The client cert must be signed by the CA the kuang server trusts
(`APP_CA_CERT_PATH`), and its CN/OU determine the agent's identity and scopes.

## Running

```bash
go run ./cmd/mcp
```

Typically it is launched by an agent host rather than run by hand. Example MCP host
configuration:

```json
{
  "mcpServers": {
    "kuang": {
      "command": "/path/to/kuang-mcp",
      "env": {
        "KUANG_URL": "https://kuang.internal:8080",
        "KUANG_CA_CERT": "/root/.jack/certs/ca.pem",
        "KUANG_CERT": "/root/.jack/certs/cert.pem",
        "KUANG_KEY": "/root/.jack/certs/key.pem"
      }
    }
  }
}
```
