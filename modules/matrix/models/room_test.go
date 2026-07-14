package models

import "testing"

func TestRoomValidate(t *testing.T) {
	r := Room{RoomID: "!abc:localhost"}
	if err := r.Validate(); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
}

func TestRoomValidateEmpty(t *testing.T) {
	r := Room{}
	if err := r.Validate(); err == nil {
		t.Fatal("expected error for empty room_id")
	}
}

func TestCreateRoomRequestValidate(t *testing.T) {
	r := CreateRoomRequest{Name: "test"}
	if err := r.Validate(); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
}

func TestCreateRoomRequestValidateEmpty(t *testing.T) {
	r := CreateRoomRequest{}
	if err := r.Validate(); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestJoinRoomRequestValidate(t *testing.T) {
	r := JoinRoomRequest{RoomIDOrAlias: "!abc:localhost"}
	if err := r.Validate(); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
}

func TestJoinRoomRequestValidateEmpty(t *testing.T) {
	r := JoinRoomRequest{}
	if err := r.Validate(); err == nil {
		t.Fatal("expected error for empty room_id_or_alias")
	}
}

func TestRoomNoValidateTypes(t *testing.T) {
	if err := (RoomInfo{}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (RoomList{}).Validate(); err != nil {
		t.Fatal(err)
	}
}
