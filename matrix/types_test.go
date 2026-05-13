package matrix

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

func TestNoValidateTypes(t *testing.T) {
	// These types have no-op validation — just ensure they don't panic.
	if err := (RoomInfo{}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (RoomList{}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (Member{}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (MemberList{}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (Message{}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (MessageList{}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (Invite{}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (InviteList{}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (WatchEvent{}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (WatchResponse{}).Validate(); err != nil {
		t.Fatal(err)
	}
}
