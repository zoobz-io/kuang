# modules/matrix

Matrix client-server tools for agents.

## Purpose

Wraps the [Matrix client-server API v3](https://spec.matrix.org/latest/client-server-api/)
against a configured homeserver and exposes it as scoped kuang tools. Each agent acts
under its own Matrix access token, resolved per request from the credential store.

Unlike a thin REST wrapper, this module **synthesizes higher-level operations** from
multiple raw Matrix calls — room enrichment, alias resolution, synthetic DMs, and a
long-poll watch loop. See [`modules/README.md`](../README.md) for the module pattern.

## Install

```go
import (
    "github.com/zoobzio/kuang/modules/matrix"
    matrixcfg "github.com/zoobzio/kuang/modules/matrix/config"
)

core.Run(matrix.Module(matrixcfg.API{Homeserver: "https://matrix.org"}))
```

## Config

| Field | Required | Notes |
|-------|----------|-------|
| `Homeserver` | yes | Base URL; the server name (host) is derived from it to build room aliases |

## Credential

`CredentialKey = "matrix-token"` — the agent's Matrix access token, sent as
`Authorization: Bearer`. Seed it per agent via `PUT /creds/matrix-token`.

## Tools

| Method + path | Scope | Behavior |
|---------------|-------|----------|
| `GET /matrix/whoami` | `matrix-identity-read` | current user |
| `GET /matrix/rooms` | `matrix-rooms-read` | joined rooms, enriched with name/topic |
| `POST /matrix/rooms` | `matrix-rooms-write` | create a room |
| `POST /matrix/rooms/join` | `matrix-rooms-write` | join by room id or alias (alias resolved) |
| `POST /matrix/rooms/{room}/leave` | `matrix-rooms-write` | leave |
| `GET /matrix/rooms/{room}/members` | `matrix-members-read` | joined members |
| `POST /matrix/rooms/{room}/messages` | `matrix-messages-write` | send a message |
| `GET /matrix/rooms/{room}/messages` | `matrix-messages-read` | read (query `limit`/`since`/`from`) |
| `POST /matrix/rooms/{room}/invite` | `matrix-invites-write` | invite a user |
| `GET /matrix/invites` | `matrix-invites-read` | pending invites |
| `POST /matrix/dm/{user}/messages` | `matrix-dm-write` | send a DM (creates the DM room if needed) |
| `GET /matrix/dm/{user}/messages` | `matrix-dm-read` | read a DM (query `limit`) |
| `GET /matrix/watch` | `matrix-watch-read` | long-poll for new messages + invites |

`handlers.Scopes()` returns all eleven scope strings.

## Notable behavior

- **DMs are synthetic.** The module reads/writes the user's `m.direct` account data,
  creating a `trusted_private_chat` room on first send.
- **`GET /matrix/watch`** is the core endpoint for agent event loops: it long-polls
  `/sync`, returns new `m.room.message` events and pending invites, and hands back a
  `next_batch` cursor to pass as `since` on the next call.
- **Reads** support filtering by sender substring (`from`) and reading since a given
  event (`since`).
