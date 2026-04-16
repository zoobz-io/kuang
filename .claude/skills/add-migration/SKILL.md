---
name: add-migration
description: Create a database migration using goose
---

# Add Migration

You are creating a database migration - a versioned SQL file that modifies the database schema. Migrations are managed by [goose](https://github.com/pressly/goose).

## Technical Foundation

Migrations live in `migrations/` as SQL files with goose annotations:

```sql
-- +goose Up
CREATE TABLE users (
    id BIGINT PRIMARY KEY,
    login TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_users_login ON users(login);

-- +goose Down
DROP TABLE users;
```

### File Naming

```
NNN_description.sql
```

- Three-digit prefix for ordering (`001`, `002`, etc.)
- Descriptive name in snake_case
- Always `.sql` extension

Examples:
- `001_create_users.sql`
- `002_create_repositories.sql`
- `003_add_user_avatar.sql`

### Goose Annotations

- `-- +goose Up` - SQL to apply the migration
- `-- +goose Down` - SQL to rollback the migration

Both sections are required in every migration file.

### Common Patterns

**Create table:**
```sql
-- +goose Up
CREATE TABLE documents (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    path TEXT NOT NULL,
    content_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, path)
);

CREATE INDEX idx_documents_user_id ON documents(user_id);

-- +goose Down
DROP TABLE documents;
```

**Add column:**
```sql
-- +goose Up
ALTER TABLE users ADD COLUMN avatar_url TEXT;

-- +goose Down
ALTER TABLE users DROP COLUMN avatar_url;
```

**Add index:**
```sql
-- +goose Up
CREATE INDEX idx_users_email ON users(email);

-- +goose Down
DROP INDEX idx_users_email;
```

**Vector column (pgvector):**
```sql
-- +goose Up
ALTER TABLE documents ADD COLUMN vector vector(1536);

-- +goose Down
ALTER TABLE documents DROP COLUMN vector;
```

**Extensions:**
```sql
-- +goose Up
CREATE EXTENSION IF NOT EXISTS vector;

-- +goose Down
DROP EXTENSION IF EXISTS vector;
```

### Running Migrations

```bash
# Via docker-compose (runs automatically on startup)
docker compose up migrate

# Manually with goose
goose -dir migrations postgres "$DATABASE_URL" up
goose -dir migrations postgres "$DATABASE_URL" down
goose -dir migrations postgres "$DATABASE_URL" status
```

## Your Task

Understand what schema change is needed:
1. Is this for a new model? (create table)
2. Is this modifying an existing table? (alter table)
3. What constraints, indexes, and foreign keys are needed?

## Before Writing Code

Produce a spec for approval:

```
## Migration: [NNN]_[description]

**Purpose:** [What this migration does]

**Up:**
[Brief description of changes]

**Down:**
[Brief description of rollback]

**Indexes:** [List any indexes to create]

**Foreign keys:** [List any FK constraints]
```

## After Approval

1. Determine the next migration number (check existing files in `migrations/`)
2. Create `migrations/NNN_description.sql` with Up and Down sections
3. Test locally: `docker compose up migrate`

## Guidelines

- Never modify existing migrations - create new ones
- Always provide a Down section
- Use `IF EXISTS` / `IF NOT EXISTS` for safety in Down section
- Keep migrations small and focused
- Include indexes in the same migration as the table
- Use `BIGSERIAL` for auto-increment primary keys
- Use `TIMESTAMPTZ` for timestamps (timezone-aware)
- Use `ON DELETE CASCADE` for child tables
