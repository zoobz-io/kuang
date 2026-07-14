package stores

import (
	"context"
	"path/filepath"
	"testing"

	// The credential store's production code opens the "sqlite3" driver but,
	// like the sibling internal/auth store, leaves driver registration to the
	// importer. The test binary registers it here.
	_ "github.com/mattn/go-sqlite3"
)

// newTestStore opens a fresh SQLite-backed credential store in a temp dir.
func newTestStore(t *testing.T) *Credentials {
	t.Helper()
	store, err := NewCredentials(filepath.Join(t.TempDir(), "creds.db"))
	if err != nil {
		t.Fatalf("NewCredentials: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func TestSetInsertsAndResolves(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	if err := store.Set(ctx, "agentA", "api-token", "secret"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	got, err := store.Resolve(ctx, "agentA", "api-token")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if want := "secret"; got != want {
		t.Errorf("Resolve got %q, want %q", got, want)
	}
}

func TestSetUpserts(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	if err := store.Set(ctx, "agentA", "api-token", "first"); err != nil {
		t.Fatalf("Set first: %v", err)
	}
	if err := store.Set(ctx, "agentA", "api-token", "second"); err != nil {
		t.Fatalf("Set second: %v", err)
	}

	// Value should be updated, not duplicated.
	got, err := store.Resolve(ctx, "agentA", "api-token")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if want := "second"; got != want {
		t.Errorf("Resolve after upsert got %q, want %q", got, want)
	}

	// The listing should show exactly one key, proving no duplicate row.
	keys, err := store.List(ctx, "agentA")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(keys) != 1 {
		t.Errorf("List got %d keys, want 1 (upsert must not duplicate): %v", len(keys), keys)
	}
}

func TestResolveUnknownKey(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	if _, err := store.Resolve(ctx, "agentA", "nope"); err == nil {
		t.Fatal("expected error resolving unknown key, got nil")
	}
}

func TestDeleteRemoves(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	if err := store.Set(ctx, "agentA", "api-token", "secret"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := store.Delete(ctx, "agentA", "api-token"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if _, err := store.Resolve(ctx, "agentA", "api-token"); err == nil {
		t.Fatal("expected error resolving deleted key, got nil")
	}
}

func TestDeleteIsScoped(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	// Both agents have a credential under the same key name.
	if err := store.Set(ctx, "agentA", "shared", "a-value"); err != nil {
		t.Fatalf("Set agentA: %v", err)
	}
	if err := store.Set(ctx, "agentB", "shared", "b-value"); err != nil {
		t.Fatalf("Set agentB: %v", err)
	}

	// agentA deleting its key must not touch agentB's same-named key.
	if err := store.Delete(ctx, "agentA", "shared"); err != nil {
		t.Fatalf("Delete agentA: %v", err)
	}

	got, err := store.Resolve(ctx, "agentB", "shared")
	if err != nil {
		t.Fatalf("Resolve agentB after agentA delete: %v", err)
	}
	if want := "b-value"; got != want {
		t.Errorf("agentB value got %q, want %q", got, want)
	}
}

func TestListScopedAndValueless(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	if err := store.Set(ctx, "agentA", "token-1", "v1"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := store.Set(ctx, "agentA", "token-2", "v2"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := store.Set(ctx, "agentB", "token-3", "v3"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	keys, err := store.List(ctx, "agentA")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("List got %d keys, want 2: %v", len(keys), keys)
	}

	// Only agentA's keys, and never any values.
	for _, k := range keys {
		if k != "token-1" && k != "token-2" {
			t.Errorf("unexpected key %q in agentA listing", k)
		}
		if k == "v1" || k == "v2" {
			t.Errorf("listing leaked a value: %q", k)
		}
		if k == "token-3" {
			t.Errorf("listing leaked another agent's key: %q", k)
		}
	}
}

func TestListEmptyForUnknownAgent(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	keys, err := store.List(ctx, "ghost")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("List for unknown agent got %d keys, want 0: %v", len(keys), keys)
	}
}

func TestAgentIsolation(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	if err := store.Set(ctx, "agentB", "secret-key", "b-only"); err != nil {
		t.Fatalf("Set agentB: %v", err)
	}

	// agentA must not be able to resolve agentB's credential.
	if _, err := store.Resolve(ctx, "agentA", "secret-key"); err == nil {
		t.Error("agentA resolved agentB's credential — isolation breach")
	}

	// agentA deleting the same key name must not remove agentB's credential.
	if err := store.Delete(ctx, "agentA", "secret-key"); err != nil {
		t.Fatalf("Delete agentA (scoped no-op): %v", err)
	}
	got, err := store.Resolve(ctx, "agentB", "secret-key")
	if err != nil {
		t.Fatalf("Resolve agentB after agentA delete attempt: %v", err)
	}
	if want := "b-only"; got != want {
		t.Errorf("agentB value got %q, want %q — isolation breach", got, want)
	}
}
