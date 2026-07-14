package models

import "testing"

func TestDMSendRequestValidate(t *testing.T) {
	r := DMSendRequest{Message: "hey"}
	if err := r.Validate(); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
}

func TestDMSendRequestValidateEmpty(t *testing.T) {
	r := DMSendRequest{}
	if err := r.Validate(); err == nil {
		t.Fatal("expected error for empty message")
	}
}
