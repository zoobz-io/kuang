# internal/mcp

A generic OpenAPI-to-MCP bridge over stdio.

## Purpose

This package implements the MCP server that an AI agent's host launches (the
`cmd/mcp` binary). It connects to a running kuang server over mTLS, fetches kuang's
OpenAPI spec, turns each operation into an MCP tool, and proxies `tools/call`
requests back to kuang as HTTP calls. Nothing here is kuang-specific beyond the
transport — it bridges any OpenAPI surface to MCP.

Protocol version: `2025-11-25`. Server identity advertised to clients:
`kuang-mcp` v`0.1.0`.

## Flow

```
agent host ──stdio(JSON-RPC)──▶ Server ──▶ Backend (Client) ──mTLS/HTTPS──▶ kuang
```

1. `NewClient(baseURL, caCert, cert, key)` builds an mTLS `http.Client` (TLS 1.3).
2. `NewServer(client)` then `LoadTools()` calls `FetchSpec` (`GET /openapi`) and
   converts each operation into a `Tool`. Because the fetch is mTLS, kuang returns a
   spec **scoped to the presenting cert** — the agent only sees tools it may call.
3. `Run(stdin, stdout)` is the newline-delimited JSON-RPC loop.

## Components

| File | Contents |
|------|----------|
| `protocol.go` | JSON-RPC 2.0 + MCP types (`Request`, `Response`, `Tool`, `ToolCallParams`, …) |
| `server.go` | `Server`, the `Backend` interface, `LoadTools`, `Run`, method dispatch |
| `client.go` | `Client` — the mTLS `Backend` impl (`FetchSpec`, `Call`) |
| `tools.go` | OpenAPI → MCP translation (`OperationsFromSpec`, `ToolFromOperation`) |

### Method dispatch (`server.go`)

`initialize` → capabilities + server info · `notifications/initialized` → no reply ·
`tools/list` → the loaded tools · `tools/call` → proxy via the backend · `ping` → `{}`
· unknown → JSON-RPC `-32601`.

### Tool translation (`tools.go`)

Each OpenAPI operation becomes one tool. The tool name is the `operationId` (or
`method_path`), sanitized. The input schema is assembled from path/query parameters
plus a required `body` property for the JSON request body.

### Call mapping (`client.go`)

`Call` reconstructs the exact HTTP request from a tool invocation: path params are
substituted into `{name}` placeholders, query params appended, and a `body` argument
JSON-marshaled as the request body. The HTTP status/body is returned to the agent as
tool result content, with `isError` set on status ≥ 400.

See [`cmd/mcp/README.md`](../../cmd/mcp/README.md) for running the bridge.
