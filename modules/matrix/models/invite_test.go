package models

import "testing"

func TestInviteRequestValidate(t *testing.T) {
	r := InviteRequest{UserID: "@user:localhost"}
	if err := r.Validate(); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
}

func TestInviteRequestValidateEmpty(t *testing.T) {
	r := InviteRequest{}
	if err := r.Validate(); err == nil {
		t.Fatal("expected error for empty user_id")
	}
}

func TestInviteNoValidateTypes(t *testing.T) {
	if err := (Invite{}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (InviteList{}).Validate(); err != nil {
		t.Fatal(err)
	}
}
