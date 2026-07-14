package auth

import (
	"context"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

// Use the cheapest bcrypt cost in tests. bcrypt at production cost is
// deliberately slow (~tens of ms/op); MinCost keeps these unit tests fast
// without changing the hashing behavior under test.
func init() { bcryptCost = bcrypt.MinCost }

// openTestStore opens a Store backed by a fresh temp-dir SQLite database.
func openTestStore(t *testing.T) *Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "admin.db")
	store, err := OpenStore(path)
	if err != nil {
		t.Fatalf("OpenStore() error: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func TestOpenStore(t *testing.T) {
	store := openTestStore(t)
	if store == nil {
		t.Fatal("expected non-nil store")
	}
}

func TestCreateAndVerify(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	if err := store.Create(ctx, "alice", "s3cret"); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	if err := store.Verify(ctx, "alice", "s3cret"); err != nil {
		t.Errorf("Verify() with correct password: got error %v, want nil", err)
	}
}

func TestVerifyWrongPassword(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	if err := store.Create(ctx, "alice", "s3cret"); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	if err := store.Verify(ctx, "alice", "wrong"); err == nil {
		t.Error("Verify() with wrong password: got nil, want error")
	}
}

func TestVerifyUnknownUser(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	// No users created. Verify against an unknown user must error
	// (the store performs a constant-time dummy compare internally).
	if err := store.Verify(ctx, "ghost", "whatever"); err == nil {
		t.Error("Verify() with unknown user: got nil, want error")
	}
}

func TestInitialized(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	got, err := store.Initialized(ctx)
	if err != nil {
		t.Fatalf("Initialized() error: %v", err)
	}
	if got {
		t.Error("Initialized() before Initialize: got true, want false")
	}

	if err := store.Initialize(ctx, "admin", "hunter2"); err != nil {
		t.Fatalf("Initialize() error: %v", err)
	}

	got, err = store.Initialized(ctx)
	if err != nil {
		t.Fatalf("Initialized() error: %v", err)
	}
	if !got {
		t.Error("Initialized() after Initialize: got false, want true")
	}
}

func TestInitializeTwice(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	if err := store.Initialize(ctx, "admin", "hunter2"); err != nil {
		t.Fatalf("first Initialize() error: %v", err)
	}

	if err := store.Initialize(ctx, "admin2", "other"); err == nil {
		t.Error("second Initialize(): got nil, want error (already initialized)")
	}
}

func TestInitializeCreatesVerifiableUser(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	if err := store.Initialize(ctx, "admin", "hunter2"); err != nil {
		t.Fatalf("Initialize() error: %v", err)
	}

	if err := store.Verify(ctx, "admin", "hunter2"); err != nil {
		t.Errorf("Verify() of initialized user: got error %v, want nil", err)
	}
}
