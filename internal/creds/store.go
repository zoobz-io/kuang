// Package creds provides kuang's internal credential store backed by SQLite.
package creds

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/zoobz-io/grub"

	astqlsqlite "github.com/zoobz-io/astql/sqlite"
)

// credential is the database model for agent credentials.
type credential struct {
	ID    string `db:"id" constraints:"primarykey"`
	Agent string `db:"agent" constraints:"notnull"`
	Key   string `db:"key" constraints:"notnull"`
	Value string `db:"value" constraints:"notnull"`
}

// Store is kuang's internal credential store backed by a grub SQLite database.
type Store struct {
	db   *grub.Database[credential]
	conn *sqlx.DB
}

// Open opens (or creates) a SQLite credential store at the given path.
func Open(path string) (*Store, error) {
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
	db := grub.NewDatabase[credential](conn, "credentials", astqlsqlite.New())
	return &Store{db: db, conn: conn}, nil
}

// Resolve returns the credential value for the given agent and key.
// This satisfies contracts.Credentials.
func (s *Store) Resolve(ctx context.Context, agent string, key string) (string, error) {
	results, err := s.db.Query().
		Where("agent", "=", "agent").
		Where("key", "=", "key").
		Exec(ctx, map[string]any{"agent": agent, "key": key})
	if err != nil {
		return "", fmt.Errorf("resolve credential: %w", err)
	}
	if len(results) == 0 {
		return "", fmt.Errorf("credential %q not found for agent %q", key, agent)
	}
	return results[0].Value, nil
}

// Set stores a credential for the given agent and key.
func (s *Store) Set(ctx context.Context, agent string, key string, value string) error {
	id := agent + ":" + key
	return s.db.Set(ctx, id, &credential{
		ID:    id,
		Agent: agent,
		Key:   key,
		Value: value,
	})
}

// Delete removes a credential for the given agent and key.
func (s *Store) Delete(ctx context.Context, agent string, key string) error {
	id := agent + ":" + key
	return s.db.Delete(ctx, id)
}

// List returns the credential keys for the given agent.
func (s *Store) List(ctx context.Context, agent string) ([]string, error) {
	results, err := s.db.Query().
		Where("agent", "=", "agent").
		Exec(ctx, map[string]any{"agent": agent})
	if err != nil {
		return nil, fmt.Errorf("list credentials: %w", err)
	}
	keys := make([]string, len(results))
	for i, r := range results {
		keys[i] = r.Key
	}
	return keys, nil
}

// Close closes the underlying database connection.
func (s *Store) Close() error {
	return s.conn.Close()
}
