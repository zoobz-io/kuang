//go:build testing

package github

import (
	"context"
	"testing"

	"github.com/zoobz-io/sum"
)

// startRegistry resets and starts the process-global sum registry for a test,
// restoring it on cleanup. Invoking a module closure registers a service, so
// each test needs a fresh started registry.
func startRegistry(t *testing.T) sum.Key {
	t.Helper()
	sum.Reset()
	t.Cleanup(sum.Reset)
	return sum.Start()
}

// TestModuleLoadsConfigFromEnv verifies the module pulls its config from the
// environment (via fig) when invoked, rather than from a constructor argument.
func TestModuleLoadsConfigFromEnv(t *testing.T) {
	t.Setenv("KUANG_GITHUB_API_URL", "https://api.github.com")
	t.Setenv("KUANG_GITHUB_OWNER", "octocat")

	k := startRegistry(t)
	eps, err := Module()(context.Background(), k)
	if err != nil {
		t.Fatalf("module invocation error: %v", err)
	}
	if len(eps) == 0 {
		t.Error("expected endpoints, got none")
	}
}

// TestModuleMissingRequiredConfig verifies that a missing required env var
// (owner has no default) fails config validation when the module is invoked.
func TestModuleMissingRequiredConfig(t *testing.T) {
	t.Setenv("KUANG_GITHUB_API_URL", "https://api.github.com")
	t.Setenv("KUANG_GITHUB_OWNER", "")

	k := startRegistry(t)
	if _, err := Module()(context.Background(), k); err == nil {
		t.Fatal("expected error when required owner is unset")
	}
}
