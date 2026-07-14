# modules/github

GitHub REST tools for agents.

## Purpose

Wraps the [GitHub REST API v3](https://docs.github.com/rest) and exposes it as scoped
kuang tools. Each agent acts under its own GitHub token (a PAT), resolved per request
from the credential store. One module instance is scoped to a single owner/org.

See [`modules/README.md`](../README.md) for the module pattern this follows.

## Install

```go
import (
    "github.com/zoobzio/kuang/modules/github"
    githubcfg "github.com/zoobzio/kuang/modules/github/config"
)

core.Run(github.Module(githubcfg.API{Owner: "zoobzio"}))
```

## Config

| Field | Required | Default | Notes |
|-------|----------|---------|-------|
| `Owner` | yes | — | Org/user used as the `{owner}` in every URL |
| `APIURL` | no | `https://api.github.com` | Upstream base URL |

The service sets `Accept: application/vnd.github+json` and
`X-GitHub-Api-Version: 2022-11-28` on every call.

## Credential

`CredentialKey = "github-token"` — the agent's GitHub PAT, sent as
`Authorization: Bearer`. Seed it per agent via `PUT /creds/github-token`.

## Tools

| Method + path | Scope | GitHub call |
|---------------|-------|-------------|
| `GET /github/repos` | `github-repos-read` | list `{owner}`'s repos |
| `GET /github/repos/{name}` | `github-repos-read` | get a repo |
| `GET /github/repos/{repo}/issues` | `github-issues-read` | list issues |
| `GET /github/repos/{repo}/issues/{number}` | `github-issues-read` | get an issue |
| `POST /github/repos/{repo}/issues` | `github-issues-write` | create an issue |
| `GET /github/repos/{repo}/pulls` | `github-pulls-read` | list PRs |
| `GET /github/repos/{repo}/pulls/{number}` | `github-pulls-read` | get a PR |
| `POST /github/repos/{repo}/pulls` | `github-pulls-write` | create a PR |
| `GET /github/repos/{repo}/content/{path...}` | `github-content-read` | get file content (query `ref`) |
| `PUT /github/repos/{repo}/content/{path...}` | `github-content-write` | create/update a file |
| `GET /github/search/code` | `github-search-read` | search code (query `query`) |

`handlers.Scopes()` returns all eight scope strings.
