# README Create

Create a README that captures the application's essence through clarity, not marketing.

## Principles

1. **Purpose over problem** — Lead with what this application does
2. **Show, don't frame** — Configuration and usage speak louder than "The Problem / The Solution"
3. **Voice from nature** — Let the application's character dictate tone
4. **Structure as scaffolding** — Sections provide rhythm, not script

## Execution

1. Read `checklist.md` in this skill directory
2. Complete all Discovery tasks before writing
3. Complete all Synthesis tasks and confirm with user
4. Write README per specifications below
5. Verify output against all requirements

## Specifications

### Badges

The README MUST include exactly these 6 badges immediately after the title. Replace `[app]` with the actual application name. DO NOT modify URLs, colors, styles, or badge providers.

| # | Badge | Markdown |
|---|-------|----------|
| 1 | CI Status | `[![CI Status](https://github.com/zoobzio/[app]/workflows/CI/badge.svg)](https://github.com/zoobzio/[app]/actions/workflows/ci.yml)` |
| 2 | Coverage | `[![codecov](https://codecov.io/gh/zoobzio/[app]/graph/badge.svg?branch=main)](https://codecov.io/gh/zoobzio/[app])` |
| 3 | CodeQL | `[![CodeQL](https://github.com/zoobzio/[app]/workflows/CodeQL/badge.svg)](https://github.com/zoobzio/[app]/security/code-scanning)` |
| 4 | Go Version | `[![Go Version](https://img.shields.io/github/go-mod/go-version/zoobzio/[app])](go.mod)` |
| 5 | License | `[![License](https://img.shields.io/github/license/zoobzio/[app])](LICENSE)` |
| 6 | Release | `[![Release](https://img.shields.io/github/v/release/zoobzio/[app])](https://github.com/zoobzio/[app]/releases)` |

### Structure

The README MUST contain these sections in this order:

| # | Section | Requirements |
|---|---------|--------------|
| 1 | Header | Title, badges (all 6), tagline + one supporting sentence |
| 2 | Overview | Application-specific name, what it does + primary use case |
| 3 | API Surfaces | Brief description of public (api/) and admin (admin/) surfaces |
| 4 | Quick Start | Prerequisites, installation, running locally |
| 5 | Configuration | Key environment variables and config options |
| 6 | Capabilities | Feature table with links to documentation |
| 7 | Architecture | Brief description with link to docs/architecture |
| 8 | Documentation | Links organized as Learn / Guides / Reference |
| 9 | Development | Setup for local development |
| 10 | Contributing | One sentence + link to CONTRIBUTING.md |
| 11 | License | Single line: "MIT License - see [LICENSE](LICENSE)" |

### Overview Section Naming

The overview section MUST have an application-specific name that captures the core purpose.

PROHIBITED names: "Overview", "Introduction", "About", "What Is This"

Examples of compliant names:
- "The Platform" (for a platform application)
- "Your Data, Your Rules" (for a data management app)
- "API Gateway" (for a gateway application)

### Code Blocks

Each code block serves a distinct purpose:

| Block | Purpose | Characteristics |
|-------|---------|-----------------|
| Quick Start | Get running | Prerequisites, commands to run |
| Configuration | Customize | Environment variables, config file examples |
| Development | Contribute | Local setup, running tests |

## Prohibitions

DO NOT:
- Use alternative badge providers or styles
- Omit any of the 6 required badges
- Use generic overview section names
- Include full API reference tables (belongs in docs/reference)
- Frame with "The Problem / The Solution"
- Write in template voice (README MUST feel unique to THIS application)

## Output

A README.md that:
- Contains all 6 badges with exact URLs specified
- Could only describe THIS application
- Opens with purpose, not marketing copy
- Provides clear navigation to deeper documentation
