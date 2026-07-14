package models

import "testing"

func TestSendMessageRequestValidate(t *testing.T) {
	r := SendMessageRequest{Message: "hello"}
	if err := r.Validate(); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
}

func TestSendMessageRequestValidateEmpty(t *testing.T) {
	r := SendMessageRequest{}
	if err := r.Validate(); err == nil {
		t.Fatal("expected error for empty message")
	}
}

func TestSendMessageResponseValidate(t *testing.T) {
	r := SendMessageResponse{EventID: "$evt123"}
	if err := r.Validate(); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
}

func TestSendMessageResponseValidateEmpty(t *testing.T) {
	r := SendMessageResponse{}
	if err := r.Validate(); err == nil {
		t.Fatal("expected error for empty event_id")
	}
}

func TestMessageNoValidateTypes(t *testing.T) {
	if err := (Message{}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (MessageList{}).Validate(); err != nil {
		t.Fatal(err)
	}
}
