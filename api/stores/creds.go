// Package stores provides kuang's internal credential store backed by SQLite.
package stores

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/zoobz-io/capitan"
	"github.com/zoobz-io/grub"
	"github.com/zoobzio/kuang/api/events"
	"github.com/zoobzio/kuang/api/models"

	astqlsqlite "github.com/zoobz-io/astql/sqlite"
)

// Credentials is kuang's internal credential store backed by a grub SQLite database.
type Credentials struct {
	db   *grub.Database[models.Credential]
	conn *sqlx.DB
}

// NewCredentials opens (or creates) a SQLite credential store at the given path.
func NewCredentials(path string) (*Credentials, error) {
	conn, err := sqlx.Connect("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("open credential store: %w", err)
	}
	if _, err := conn.Exec(`CREATE TABLE IF NOT EXISTS credentials (
		id    TEXT PRIMARY KEY,
		agent TEXT NOT NULL,
		key   TEXT NOT NULL,
		value TEXT NOT NULL,
		UNIQUE(agent, key)
	)`); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("init credential store: %w", err)
	}
	db := grub.NewDatabase[models.Credential](conn, "credentials", astqlsqlite.New())
	return &Credentials{db: db, conn: conn}, nil
}

// Resolve returns the credential value for the given agent and key.
// This satisfies contracts.Credentials.
func (s *Credentials) Resolve(ctx context.Context, agent string, key string) (string, error) {
	results, err := s.db.Query().
		Where("agent", "=", "agent").
		Where("key", "=", "key").
		Exec(ctx, map[string]any{"agent": agent, "key": key})
	if err != nil {
		capitan.Error(ctx, events.CredentialFailed,
			events.CredentialOperationKey.Field("resolve"),
			events.CredentialAgentKey.Field(agent),
			events.CredentialKeyNameKey.Field(key),
			events.CredentialErrorKey.Field(err.Error()),
		)
		return "", fmt.Errorf("resolve credential: %w", err)
	}
	if len(results) == 0 {
		capitan.Warn(ctx, events.CredentialNotFound,
			events.CredentialAgentKey.Field(agent),
			events.CredentialKeyNameKey.Field(key),
		)
		return "", fmt.Errorf("credential %q not found for agent %q", key, agent)
	}
	capitan.Info(ctx, events.CredentialResolved,
		events.CredentialAgentKey.Field(agent),
		events.CredentialKeyNameKey.Field(key),
	)
	return results[0].Value, nil
}

// Set creates or updates a credential for the given agent and key. The upsert
// targets the (agent, key) unique constraint, so a caller-controlled key can
// never collide with a different agent's row. The id is an opaque surrogate
// key used only on insert; on conflict the existing row's id is kept.
func (s *Credentials) Set(ctx context.Context, agent string, key string, value string) error {
	_, err := s.db.InsertFull().
		OnConflict("agent", "key").
		DoUpdate().
		Set("value", "value").
		Exec(ctx, &models.Credential{
			ID:    uuid.NewString(),
			Agent: agent,
			Key:   key,
			Value: value,
		})
	if err != nil {
		capitan.Error(ctx, events.CredentialFailed,
			events.CredentialOperationKey.Field("set"),
			events.CredentialAgentKey.Field(agent),
			events.CredentialKeyNameKey.Field(key),
			events.CredentialErrorKey.Field(err.Error()),
		)
		return fmt.Errorf("set credential: %w", err)
	}
	capitan.Info(ctx, events.CredentialSet,
		events.CredentialAgentKey.Field(agent),
		events.CredentialKeyNameKey.Field(key),
	)
	return nil
}

// Delete removes a credential for the given agent and key. The delete is scoped
// by both columns, so a caller cannot remove another agent's credential.
func (s *Credentials) Delete(ctx context.Context, agent string, key string) error {
	if _, err := s.db.Remove().
		Where("agent", "=", "agent").
		Where("key", "=", "key").
		Exec(ctx, map[string]any{"agent": agent, "key": key}); err != nil {
		capitan.Error(ctx, events.CredentialFailed,
			events.CredentialOperationKey.Field("delete"),
			events.CredentialAgentKey.Field(agent),
			events.CredentialKeyNameKey.Field(key),
			events.CredentialErrorKey.Field(err.Error()),
		)
		return fmt.Errorf("delete credential: %w", err)
	}
	capitan.Info(ctx, events.CredentialDeleted,
		events.CredentialAgentKey.Field(agent),
		events.CredentialKeyNameKey.Field(key),
	)
	return nil
}

// List returns the credential keys for the given agent.
func (s *Credentials) List(ctx context.Context, agent string) ([]string, error) {
	results, err := s.db.Query().
		Where("agent", "=", "agent").
		Exec(ctx, map[string]any{"agent": agent})
	if err != nil {
		capitan.Error(ctx, events.CredentialFailed,
			events.CredentialOperationKey.Field("list"),
			events.CredentialAgentKey.Field(agent),
			events.CredentialErrorKey.Field(err.Error()),
		)
		return nil, fmt.Errorf("list credentials: %w", err)
	}
	keys := make([]string, len(results))
	for i, r := range results {
		keys[i] = r.Key
	}
	capitan.Info(ctx, events.CredentialListed,
		events.CredentialAgentKey.Field(agent),
		events.CredentialCountKey.Field(len(keys)),
	)
	return keys, nil
}

// Close closes the underlying database connection.
func (s *Credentials) Close() error {
	return s.conn.Close()
}
