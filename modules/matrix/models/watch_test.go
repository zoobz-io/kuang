package models

import "testing"

func TestWatchNoValidateTypes(t *testing.T) {
	if err := (WatchEvent{}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (WatchResponse{}).Validate(); err != nil {
		t.Fatal(err)
	}
}
