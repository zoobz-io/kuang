package models

import "testing"

func TestIdentityValidate(t *testing.T) {
	i := Identity{UserID: "@bot:localhost"}
	if err := i.Validate(); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
}

func TestIdentityValidateEmpty(t *testing.T) {
	i := Identity{}
	if err := i.Validate(); err == nil {
		t.Fatal("expected error for empty user_id")
	}
}
