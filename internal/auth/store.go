// Package auth provides authentication for the kuang admin API.
package auth

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/zoobz-io/grub"
	"golang.org/x/crypto/bcrypt"

	astqlsqlite "github.com/zoobz-io/astql/sqlite"
)

// bcryptCost is the cost used when hashing admin passwords. It defaults to
// bcrypt.DefaultCost in production; tests lower it to bcrypt.MinCost so the
// suite stays fast (bcrypt is deliberately expensive at production cost).
var bcryptCost = bcrypt.DefaultCost

// user is the database model for admin users.
type user struct {
	ID       string `db:"id" constraints:"primarykey"`
	Username string `db:"username" constraints:"notnull,unique"`
	Password string `db:"password" constraints:"notnull"`
}

// Store manages admin user credentials backed by a grub SQLite database.
type Store struct {
	db   *grub.Database[user]
	conn *sqlx.DB
}

// OpenStore opens (or creates) a SQLite user store at the given path.
func OpenStore(path string) (*Store, error) {
	conn, err := sqlx.Connect("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("open user store: %w", err)
	}
	if _, err := conn.Exec(`CREATE TABLE IF NOT EXISTS users (
		id       TEXT PRIMARY KEY,
		username TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS meta (
		key   TEXT PRIMARY KEY,
		value TEXT NOT NULL
	)`); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("init user store: %w", err)
	}
	db := grub.NewDatabase[user](conn, "users", astqlsqlite.New())
	return &Store{db: db, conn: conn}, nil
}

// Create adds a new admin user with a bcrypt-hashed password.
func (s *Store) Create(ctx context.Context, username, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	return s.db.Set(ctx, username, &user{
		ID:       username,
		Username: username,
		Password: string(hash),
	})
}

// Verify checks a username/password pair against the store.
// Returns nil on success, an error if the credentials are invalid.
func (s *Store) Verify(ctx context.Context, username, password string) error {
	results, err := s.db.Query().
		Where("username", "=", "username").
		Exec(ctx, map[string]any{"username": username})
	if err != nil {
		return fmt.Errorf("lookup user: %w", err)
	}
	if len(results) == 0 {
		// Constant-time compare against a dummy hash to prevent timing attacks.
		_ = bcrypt.CompareHashAndPassword([]byte("$2a$10$000000000000000000000u"), []byte(password))
		return fmt.Errorf("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(results[0].Password), []byte(password)); err != nil {
		return fmt.Errorf("invalid credentials")
	}
	return nil
}

// Initialized returns true if first-time setup has been completed.
func (s *Store) Initialized(ctx context.Context) (bool, error) {
	var count int
	if err := s.conn.GetContext(ctx, &count, "SELECT COUNT(*) FROM meta WHERE key = 'initialized'"); err != nil {
		return false, fmt.Errorf("check initialized: %w", err)
	}
	return count > 0, nil
}

// Initialize creates the first admin user and marks the store as initialized
// in a single transaction. Returns an error if already initialized.
func (s *Store) Initialize(ctx context.Context, username, password string) error {
	tx, err := s.conn.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var count int
	if txErr := tx.GetContext(ctx, &count, "SELECT COUNT(*) FROM meta WHERE key = 'initialized'"); txErr != nil {
		return fmt.Errorf("check initialized: %w", txErr)
	}
	if count > 0 {
		return fmt.Errorf("already initialized")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	if _, err := tx.ExecContext(ctx, "INSERT INTO users (id, username, password) VALUES (?, ?, ?)", username, username, string(hash)); err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO meta (key, value) VALUES ('initialized', '1')"); err != nil {
		return fmt.Errorf("set initialized: %w", err)
	}

	return tx.Commit()
}

// Close closes the underlying database connection.
func (s *Store) Close() error {
	return s.conn.Close()
}
