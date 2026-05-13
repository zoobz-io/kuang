package matrix

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/zoobz-io/kuang/httpc"
)

// testService creates a service backed by the given test server.
func testService(ts *httptest.Server) *service {
	return &service{
		client:     httpc.New(httpc.WithBaseURL(ts.URL)),
		homeserver: ts.URL,
	}
}

// jsonHandler returns an http.HandlerFunc that responds with the given value as JSON.
func jsonHandler(v any) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(v)
	}
}

// mux helpers — register paths and return the test server.
func newTestServer(routes map[string]http.HandlerFunc) *httptest.Server {
	mux := http.NewServeMux()
	for pattern, handler := range routes {
		mux.HandleFunc(pattern, handler)
	}
	return httptest.NewServer(mux)
}

// --- Identity ---

func TestWhoAmI(t *testing.T) {
	ts := newTestServer(map[string]http.HandlerFunc{
		"GET /_matrix/client/v3/account/whoami": jsonHandler(map[string]string{
			"user_id": "@bot:localhost",
		}),
	})
	defer ts.Close()

	svc := testService(ts)
	id, err := svc.WhoAmI(context.Background())
	if err != nil {
		t.Fatalf("WhoAmI: %v", err)
	}
	if id.UserID != "@bot:localhost" {
		t.Errorf("user_id = %q, want @bot:localhost", id.UserID)
	}
}

// --- Rooms ---

func TestCreateRoom(t *testing.T) {
	ts := newTestServer(map[string]http.HandlerFunc{
		"POST /_matrix/client/v3/createRoom": jsonHandler(map[string]string{
			"room_id": "!new:localhost",
		}),
	})
	defer ts.Close()

	svc := testService(ts)
	room, err := svc.CreateRoom(context.Background(), "test-room", "a topic", "")
	if err != nil {
		t.Fatalf("CreateRoom: %v", err)
	}
	if room.RoomID != "!new:localhost" {
		t.Errorf("room_id = %q, want !new:localhost", room.RoomID)
	}
}

func TestCreateRoomWithAlias(t *testing.T) {
	var gotBody map[string]any
	ts := newTestServer(map[string]http.HandlerFunc{
		"POST /_matrix/client/v3/createRoom": func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewDecoder(r.Body).Decode(&gotBody)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"room_id": "!aliased:localhost"})
		},
	})
	defer ts.Close()

	svc := testService(ts)
	room, err := svc.CreateRoom(context.Background(), "public-room", "", "general")
	if err != nil {
		t.Fatalf("CreateRoom with alias: %v", err)
	}
	if room.RoomID != "!aliased:localhost" {
		t.Errorf("room_id = %q", room.RoomID)
	}
	if gotBody["preset"] != "public_chat" {
		t.Errorf("preset = %v, want public_chat", gotBody["preset"])
	}
	if gotBody["room_alias_name"] != "general" {
		t.Errorf("room_alias_name = %v, want general", gotBody["room_alias_name"])
	}
}

func TestJoinedRooms(t *testing.T) {
	ts := newTestServer(map[string]http.HandlerFunc{
		"GET /_matrix/client/v3/joined_rooms": jsonHandler(map[string]any{
			"joined_rooms": []string{"!a:localhost", "!b:localhost"},
		}),
		// getRoomInfo calls — return 404 so they're gracefully skipped.
		"/": func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "not found", http.StatusNotFound)
		},
	})
	defer ts.Close()

	svc := testService(ts)
	list, err := svc.JoinedRooms(context.Background())
	if err != nil {
		t.Fatalf("JoinedRooms: %v", err)
	}
	if len(list.Rooms) != 2 {
		t.Fatalf("rooms = %d, want 2", len(list.Rooms))
	}
	if list.Rooms[0].RoomID != "!a:localhost" {
		t.Errorf("rooms[0].room_id = %q", list.Rooms[0].RoomID)
	}
}

func TestJoinByID(t *testing.T) {
	ts := newTestServer(map[string]http.HandlerFunc{
		"/": func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" && strings.Contains(r.URL.Path, "/join/") {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]string{"room_id": "!abc:localhost"})
				return
			}
			http.Error(w, "not found", http.StatusNotFound)
		},
	})
	defer ts.Close()

	svc := testService(ts)
	room, err := svc.Join(context.Background(), "!abc:localhost")
	if err != nil {
		t.Fatalf("Join: %v", err)
	}
	if room.RoomID != "!abc:localhost" {
		t.Errorf("room_id = %q", room.RoomID)
	}
}

func TestJoinByAlias(t *testing.T) {
	ts := newTestServer(map[string]http.HandlerFunc{
		"/": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if strings.Contains(r.URL.Path, "/directory/room/") {
				_ = json.NewEncoder(w).Encode(map[string]string{"room_id": "!resolved:localhost"})
				return
			}
			if strings.Contains(r.URL.Path, "/join/") {
				_ = json.NewEncoder(w).Encode(map[string]string{"room_id": "!resolved:localhost"})
				return
			}
			http.Error(w, "not found", http.StatusNotFound)
		},
	})
	defer ts.Close()

	svc := testService(ts)
	room, err := svc.Join(context.Background(), "#general:localhost")
	if err != nil {
		t.Fatalf("Join alias: %v", err)
	}
	if room.RoomID != "!resolved:localhost" {
		t.Errorf("room_id = %q", room.RoomID)
	}
}

func TestLeave(t *testing.T) {
	ts := newTestServer(map[string]http.HandlerFunc{
		"/": func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" && strings.Contains(r.URL.Path, "/leave") {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{}`))
				return
			}
			http.Error(w, "not found", http.StatusNotFound)
		},
	})
	defer ts.Close()

	svc := testService(ts)
	if err := svc.Leave(context.Background(), "!abc:localhost"); err != nil {
		t.Fatalf("Leave: %v", err)
	}
}

// --- Members ---

func TestMembers(t *testing.T) {
	ts := newTestServer(map[string]http.HandlerFunc{
		"/": func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"joined": map[string]any{
					"@alice:localhost": map[string]string{"display_name": "Alice"},
					"@bob:localhost":   map[string]string{"display_name": "Bob"},
				},
			})
		},
	})
	defer ts.Close()

	svc := testService(ts)
	list, err := svc.Members(context.Background(), "!room:localhost")
	if err != nil {
		t.Fatalf("Members: %v", err)
	}
	if len(list.Members) != 2 {
		t.Fatalf("members = %d, want 2", len(list.Members))
	}
}

// --- Messages ---

func TestSend(t *testing.T) {
	ts := newTestServer(map[string]http.HandlerFunc{
		"/": func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "PUT" && strings.Contains(r.URL.Path, "/send/") {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]string{"event_id": "$evt1"})
				return
			}
			http.Error(w, "not found", http.StatusNotFound)
		},
	})
	defer ts.Close()

	svc := testService(ts)
	resp, err := svc.Send(context.Background(), "!room:localhost", "hello")
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if resp.EventID != "$evt1" {
		t.Errorf("event_id = %q, want $evt1", resp.EventID)
	}
}

func TestReadMessages(t *testing.T) {
	ts := newTestServer(map[string]http.HandlerFunc{
		"/": func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"chunk": []map[string]any{
					{"sender": "@bob:localhost", "type": "m.room.message", "content": map[string]any{"body": "third"}, "event_id": "$3"},
					{"sender": "@alice:localhost", "type": "m.room.message", "content": map[string]any{"body": "second"}, "event_id": "$2"},
					{"sender": "@bob:localhost", "type": "m.room.message", "content": map[string]any{"body": "first"}, "event_id": "$1"},
				},
			})
		},
	})
	defer ts.Close()

	svc := testService(ts)
	list, err := svc.ReadMessages(context.Background(), "!room:localhost", 10, "")
	if err != nil {
		t.Fatalf("ReadMessages: %v", err)
	}
	if len(list.Messages) != 3 {
		t.Fatalf("messages = %d, want 3", len(list.Messages))
	}
	// Reversed — oldest first.
	if list.Messages[0].Body != "first" {
		t.Errorf("messages[0].body = %q, want first", list.Messages[0].Body)
	}
	if list.Messages[2].Body != "third" {
		t.Errorf("messages[2].body = %q, want third", list.Messages[2].Body)
	}
}

func TestReadMessagesFilterBySender(t *testing.T) {
	ts := newTestServer(map[string]http.HandlerFunc{
		"/": func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"chunk": []map[string]any{
					{"sender": "@bob:localhost", "type": "m.room.message", "content": map[string]any{"body": "from bob"}, "event_id": "$2"},
					{"sender": "@alice:localhost", "type": "m.room.message", "content": map[string]any{"body": "from alice"}, "event_id": "$1"},
				},
			})
		},
	})
	defer ts.Close()

	svc := testService(ts)
	list, err := svc.ReadMessages(context.Background(), "!room:localhost", 10, "alice")
	if err != nil {
		t.Fatalf("ReadMessages with from: %v", err)
	}
	if len(list.Messages) != 1 {
		t.Fatalf("messages = %d, want 1", len(list.Messages))
	}
	if list.Messages[0].Sender != "@alice:localhost" {
		t.Errorf("sender = %q, want @alice:localhost", list.Messages[0].Sender)
	}
}

func TestReadMessagesSkipsNonMessageEvents(t *testing.T) {
	ts := newTestServer(map[string]http.HandlerFunc{
		"/": func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"chunk": []map[string]any{
					{"sender": "@alice:localhost", "type": "m.room.member", "content": map[string]any{}, "event_id": "$1"},
					{"sender": "@bob:localhost", "type": "m.room.message", "content": map[string]any{"body": "hello"}, "event_id": "$2"},
				},
			})
		},
	})
	defer ts.Close()

	svc := testService(ts)
	list, err := svc.ReadMessages(context.Background(), "!room:localhost", 10, "")
	if err != nil {
		t.Fatalf("ReadMessages: %v", err)
	}
	if len(list.Messages) != 1 {
		t.Fatalf("messages = %d, want 1 (should skip m.room.member)", len(list.Messages))
	}
}

func TestReadMessagesSince(t *testing.T) {
	callCount := 0
	ts := newTestServer(map[string]http.HandlerFunc{
		"/": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if strings.Contains(r.URL.Path, "/context/") {
				_ = json.NewEncoder(w).Encode(map[string]string{"end": "t100"})
				return
			}
			if strings.Contains(r.URL.Path, "/messages") {
				callCount++
				// Verify forward direction.
				if !strings.Contains(r.URL.RawQuery, "dir=f") {
					http.Error(w, "expected dir=f", http.StatusBadRequest)
					return
				}
				_ = json.NewEncoder(w).Encode(map[string]any{
					"chunk": []map[string]any{
						{"sender": "@alice:localhost", "type": "m.room.message", "content": map[string]any{"body": "new msg"}, "event_id": "$new"},
					},
				})
				return
			}
			http.Error(w, "not found", http.StatusNotFound)
		},
	})
	defer ts.Close()

	svc := testService(ts)
	list, err := svc.ReadMessagesSince(context.Background(), "!room:localhost", "$old", 10, "")
	if err != nil {
		t.Fatalf("ReadMessagesSince: %v", err)
	}
	if len(list.Messages) != 1 {
		t.Fatalf("messages = %d, want 1", len(list.Messages))
	}
	if list.Messages[0].Body != "new msg" {
		t.Errorf("body = %q, want new msg", list.Messages[0].Body)
	}
}

func TestReadMessagesSinceWithFromFilter(t *testing.T) {
	ts := newTestServer(map[string]http.HandlerFunc{
		"/": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if strings.Contains(r.URL.Path, "/context/") {
				_ = json.NewEncoder(w).Encode(map[string]string{"end": "t100"})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"chunk": []map[string]any{
					{"sender": "@alice:localhost", "type": "m.room.message", "content": map[string]any{"body": "alice msg"}, "event_id": "$a"},
					{"sender": "@bob:localhost", "type": "m.room.message", "content": map[string]any{"body": "bob msg"}, "event_id": "$b"},
				},
			})
		},
	})
	defer ts.Close()

	svc := testService(ts)
	list, err := svc.ReadMessagesSince(context.Background(), "!room:localhost", "$old", 10, "bob")
	if err != nil {
		t.Fatalf("ReadMessagesSince with from: %v", err)
	}
	if len(list.Messages) != 1 {
		t.Fatalf("messages = %d, want 1", len(list.Messages))
	}
	if list.Messages[0].Sender != "@bob:localhost" {
		t.Errorf("sender = %q, want @bob:localhost", list.Messages[0].Sender)
	}
}

// --- Invites ---

func TestInvite(t *testing.T) {
	var gotBody map[string]any
	ts := newTestServer(map[string]http.HandlerFunc{
		"/": func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" && strings.Contains(r.URL.Path, "/invite") {
				_ = json.NewDecoder(r.Body).Decode(&gotBody)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{}`))
				return
			}
			http.Error(w, "not found", http.StatusNotFound)
		},
	})
	defer ts.Close()

	svc := testService(ts)
	if err := svc.Invite(context.Background(), "!room:localhost", "@user:localhost"); err != nil {
		t.Fatalf("Invite: %v", err)
	}
	if gotBody["user_id"] != "@user:localhost" {
		t.Errorf("user_id = %v", gotBody["user_id"])
	}
}

func TestListInvites(t *testing.T) {
	ts := newTestServer(map[string]http.HandlerFunc{
		"/": func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"next_batch": "batch1",
				"rooms": map[string]any{
					"join": map[string]any{},
					"invite": map[string]any{
						"!inv:localhost": map[string]any{
							"invite_state": map[string]any{
								"events": []map[string]any{
									{"type": "m.room.name", "content": map[string]any{"name": "Cool Room"}, "sender": "@admin:localhost"},
									{"type": "m.room.member", "content": map[string]any{"membership": "invite"}, "sender": "@admin:localhost"},
								},
							},
						},
					},
				},
			})
		},
	})
	defer ts.Close()

	svc := testService(ts)
	list, err := svc.ListInvites(context.Background())
	if err != nil {
		t.Fatalf("ListInvites: %v", err)
	}
	if len(list.Invites) != 1 {
		t.Fatalf("invites = %d, want 1", len(list.Invites))
	}
	if list.Invites[0].Name != "Cool Room" {
		t.Errorf("name = %q, want Cool Room", list.Invites[0].Name)
	}
	if list.Invites[0].Sender != "@admin:localhost" {
		t.Errorf("sender = %q, want @admin:localhost", list.Invites[0].Sender)
	}
}

// --- DMs ---

func TestDMSendExistingRoom(t *testing.T) {
	ts := newTestServer(map[string]http.HandlerFunc{
		"GET /_matrix/client/v3/account/whoami": jsonHandler(map[string]string{"user_id": "@me:localhost"}),
		"/": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if strings.Contains(r.URL.Path, "/account_data/m.direct") && r.Method == "GET" {
				_ = json.NewEncoder(w).Encode(map[string][]string{
					"@friend:localhost": {"!dm:localhost"},
				})
				return
			}
			if r.Method == "PUT" && strings.Contains(r.URL.Path, "/send/") {
				_ = json.NewEncoder(w).Encode(map[string]string{"event_id": "$dm1"})
				return
			}
			http.Error(w, "not found", http.StatusNotFound)
		},
	})
	defer ts.Close()

	svc := testService(ts)
	resp, err := svc.DMSend(context.Background(), "@friend:localhost", "hey")
	if err != nil {
		t.Fatalf("DMSend: %v", err)
	}
	if resp.EventID != "$dm1" {
		t.Errorf("event_id = %q, want $dm1", resp.EventID)
	}
}

func TestDMSendCreatesRoom(t *testing.T) {
	createCalled := false
	ts := newTestServer(map[string]http.HandlerFunc{
		"GET /_matrix/client/v3/account/whoami": jsonHandler(map[string]string{"user_id": "@me:localhost"}),
		"/": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if strings.Contains(r.URL.Path, "/account_data/m.direct") {
				if r.Method == "GET" {
					_ = json.NewEncoder(w).Encode(map[string][]string{})
					return
				}
				// PUT — set direct rooms.
				_, _ = w.Write([]byte(`{}`))
				return
			}
			if r.Method == "POST" && r.URL.Path == "/_matrix/client/v3/createRoom" {
				createCalled = true
				_ = json.NewEncoder(w).Encode(map[string]string{"room_id": "!newdm:localhost"})
				return
			}
			if r.Method == "PUT" && strings.Contains(r.URL.Path, "/send/") {
				_ = json.NewEncoder(w).Encode(map[string]string{"event_id": "$dm2"})
				return
			}
			http.Error(w, "not found", http.StatusNotFound)
		},
	})
	defer ts.Close()

	svc := testService(ts)
	resp, err := svc.DMSend(context.Background(), "@new:localhost", "hello")
	if err != nil {
		t.Fatalf("DMSend new: %v", err)
	}
	if !createCalled {
		t.Error("expected createRoom to be called")
	}
	if resp.EventID != "$dm2" {
		t.Errorf("event_id = %q", resp.EventID)
	}
}

func TestDMRead(t *testing.T) {
	ts := newTestServer(map[string]http.HandlerFunc{
		"GET /_matrix/client/v3/account/whoami": jsonHandler(map[string]string{"user_id": "@me:localhost"}),
		"/": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if strings.Contains(r.URL.Path, "/account_data/m.direct") {
				_ = json.NewEncoder(w).Encode(map[string][]string{
					"@friend:localhost": {"!dm:localhost"},
				})
				return
			}
			if strings.Contains(r.URL.Path, "/messages") {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"chunk": []map[string]any{
						{"sender": "@friend:localhost", "type": "m.room.message", "content": map[string]any{"body": "hi"}, "event_id": "$1"},
					},
				})
				return
			}
			http.Error(w, "not found", http.StatusNotFound)
		},
	})
	defer ts.Close()

	svc := testService(ts)
	list, err := svc.DMRead(context.Background(), "@friend:localhost", 10)
	if err != nil {
		t.Fatalf("DMRead: %v", err)
	}
	if len(list.Messages) != 1 {
		t.Fatalf("messages = %d, want 1", len(list.Messages))
	}
}

func TestDMReadNoDMRoom(t *testing.T) {
	ts := newTestServer(map[string]http.HandlerFunc{
		"GET /_matrix/client/v3/account/whoami": jsonHandler(map[string]string{"user_id": "@me:localhost"}),
		"/": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if strings.Contains(r.URL.Path, "/account_data/m.direct") {
				_ = json.NewEncoder(w).Encode(map[string][]string{})
				return
			}
			http.Error(w, "not found", http.StatusNotFound)
		},
	})
	defer ts.Close()

	svc := testService(ts)
	_, err := svc.DMRead(context.Background(), "@stranger:localhost", 10)
	if err == nil {
		t.Fatal("expected error for missing DM room")
	}
	if !strings.Contains(err.Error(), "no DM room") {
		t.Errorf("error = %q, want 'no DM room'", err.Error())
	}
}

// --- Watch ---

func TestWatchWithMessages(t *testing.T) {
	callNum := 0
	ts := newTestServer(map[string]http.HandlerFunc{
		"/": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			callNum++
			if callNum == 1 {
				// Initial sync.
				_ = json.NewEncoder(w).Encode(map[string]any{
					"next_batch": "batch1",
					"rooms":      map[string]any{"join": map[string]any{}, "invite": map[string]any{}},
				})
				return
			}
			// Second sync — return messages.
			_ = json.NewEncoder(w).Encode(map[string]any{
				"next_batch": "batch2",
				"rooms": map[string]any{
					"join": map[string]any{
						"!room:localhost": map[string]any{
							"timeline": map[string]any{
								"events": []map[string]any{
									{"sender": "@alice:localhost", "type": "m.room.message", "content": map[string]any{"body": "hello agent"}, "event_id": "$msg1"},
								},
							},
						},
					},
					"invite": map[string]any{},
				},
			})
		},
	})
	defer ts.Close()

	svc := testService(ts)
	resp, err := svc.Watch(context.Background(), "", 1)
	if err != nil {
		t.Fatalf("Watch: %v", err)
	}
	if resp.NextBatch != "batch2" {
		t.Errorf("next_batch = %q, want batch2", resp.NextBatch)
	}
	if len(resp.Events) != 1 {
		t.Fatalf("events = %d, want 1", len(resp.Events))
	}
	ev := resp.Events[0]
	if ev.Type != "message" {
		t.Errorf("type = %q, want message", ev.Type)
	}
	if ev.Body != "hello agent" {
		t.Errorf("body = %q, want hello agent", ev.Body)
	}
	if ev.Sender != "@alice:localhost" {
		t.Errorf("sender = %q", ev.Sender)
	}
	if ev.RoomID != "!room:localhost" {
		t.Errorf("room_id = %q", ev.RoomID)
	}
}

func TestWatchWithSinceToken(t *testing.T) {
	ts := newTestServer(map[string]http.HandlerFunc{
		"/": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			// Should skip initial sync when since is provided.
			if !strings.Contains(r.URL.RawQuery, "since=existing_token") {
				t.Error("expected since=existing_token in query")
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"next_batch": "batch3",
				"rooms":      map[string]any{"join": map[string]any{}, "invite": map[string]any{}},
			})
		},
	})
	defer ts.Close()

	svc := testService(ts)
	resp, err := svc.Watch(context.Background(), "existing_token", 1)
	if err != nil {
		t.Fatalf("Watch with since: %v", err)
	}
	if resp.NextBatch != "batch3" {
		t.Errorf("next_batch = %q", resp.NextBatch)
	}
	if len(resp.Events) != 0 {
		t.Errorf("events = %d, want 0", len(resp.Events))
	}
}

func TestWatchWithInvites(t *testing.T) {
	callNum := 0
	ts := newTestServer(map[string]http.HandlerFunc{
		"/": func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			callNum++
			if callNum == 1 {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"next_batch": "b1",
					"rooms":      map[string]any{"join": map[string]any{}, "invite": map[string]any{}},
				})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"next_batch": "b2",
				"rooms": map[string]any{
					"join": map[string]any{},
					"invite": map[string]any{
						"!inv:localhost": map[string]any{
							"invite_state": map[string]any{
								"events": []map[string]any{
									{"type": "m.room.name", "content": map[string]any{"name": "Secret Room"}, "sender": "@host:localhost"},
									{"type": "m.room.member", "content": map[string]any{"membership": "invite"}, "sender": "@host:localhost"},
								},
							},
						},
					},
				},
			})
		},
	})
	defer ts.Close()

	svc := testService(ts)
	resp, err := svc.Watch(context.Background(), "", 1)
	if err != nil {
		t.Fatalf("Watch invites: %v", err)
	}
	if len(resp.Events) != 1 {
		t.Fatalf("events = %d, want 1", len(resp.Events))
	}
	ev := resp.Events[0]
	if ev.Type != "invite" {
		t.Errorf("type = %q, want invite", ev.Type)
	}
	if ev.RoomName != "Secret Room" {
		t.Errorf("room_name = %q, want Secret Room", ev.RoomName)
	}
	if ev.Sender != "@host:localhost" {
		t.Errorf("sender = %q", ev.Sender)
	}
}

func TestWatchSkipsNonMessageEvents(t *testing.T) {
	callNum := 0
	ts := newTestServer(map[string]http.HandlerFunc{
		"/": func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			callNum++
			if callNum == 1 {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"next_batch": "b1",
					"rooms":      map[string]any{"join": map[string]any{}, "invite": map[string]any{}},
				})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"next_batch": "b2",
				"rooms": map[string]any{
					"join": map[string]any{
						"!room:localhost": map[string]any{
							"timeline": map[string]any{
								"events": []map[string]any{
									{"sender": "@alice:localhost", "type": "m.room.member", "content": map[string]any{}, "event_id": "$join"},
									{"sender": "@alice:localhost", "type": "m.room.message", "content": map[string]any{"body": "actual message"}, "event_id": "$msg"},
								},
							},
						},
					},
					"invite": map[string]any{},
				},
			})
		},
	})
	defer ts.Close()

	svc := testService(ts)
	resp, err := svc.Watch(context.Background(), "", 1)
	if err != nil {
		t.Fatalf("Watch: %v", err)
	}
	if len(resp.Events) != 1 {
		t.Fatalf("events = %d, want 1 (should skip m.room.member)", len(resp.Events))
	}
	if resp.Events[0].Body != "actual message" {
		t.Errorf("body = %q", resp.Events[0].Body)
	}
}

// --- Construction ---

func TestNewService(t *testing.T) {
	svc := newService(Config{
		Homeserver:  "https://matrix.example.com/",
		AccessToken: "syt_test",
	})
	if svc.homeserver != "https://matrix.example.com" {
		t.Errorf("homeserver = %q, want trailing slash trimmed", svc.homeserver)
	}
	if svc.client == nil {
		t.Error("expected non-nil client")
	}
}

// --- Helpers ---

func TestEscapePathParam(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"!room:localhost", "%21room%3Alocalhost"},
		{"#alias:example.com", "%23alias%3Aexample.com"},
		{"simple", "simple"},
	}
	for _, tt := range tests {
		got := escapePathParam(tt.in)
		if got != tt.want {
			t.Errorf("escapePathParam(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestSenderMatches(t *testing.T) {
	tests := []struct {
		sender, filter string
		want           bool
	}{
		{"@alice:localhost", "alice", true},
		{"@alice:localhost", "@alice", true},
		{"@alice:localhost", "bob", false},
		{"@alice:localhost", "@alice:localhost", true},
		{"@bob:example.com", "example", true},
	}
	for _, tt := range tests {
		got := senderMatches(tt.sender, tt.filter)
		if got != tt.want {
			t.Errorf("senderMatches(%q, %q) = %v, want %v", tt.sender, tt.filter, got, tt.want)
		}
	}
}

func TestExtractMessage(t *testing.T) {
	msg := rawEvent{
		Sender:  "@alice:localhost",
		Type:    "m.room.message",
		Content: map[string]any{"body": "hello"},
		EventID: "$1",
	}

	m, ok := extractMessage(msg, "")
	if !ok {
		t.Fatal("expected ok")
	}
	if m.Body != "hello" {
		t.Errorf("body = %q", m.Body)
	}

	// With matching filter.
	m, ok = extractMessage(msg, "alice")
	if !ok {
		t.Fatal("expected ok with matching filter")
	}

	// With non-matching filter.
	_, ok = extractMessage(msg, "bob")
	if ok {
		t.Fatal("expected not ok with non-matching filter")
	}

	// Non-message type.
	nonMsg := rawEvent{Type: "m.room.member"}
	_, ok = extractMessage(nonMsg, "")
	if ok {
		t.Fatal("expected not ok for non-message event")
	}
}

func TestParseInvites(t *testing.T) {
	resp := &syncResponse{
		Rooms: syncRooms{
			Invite: map[string]syncInvitedRoom{
				"!inv1:localhost": {
					InviteState: syncInviteState{
						Events: []syncEvent{
							{Type: "m.room.name", Content: map[string]any{"name": "Room A"}, Sender: "@admin:localhost"},
							{Type: "m.room.member", Content: map[string]any{"membership": "invite"}, Sender: "@admin:localhost"},
						},
					},
				},
				"!inv2:localhost": {
					InviteState: syncInviteState{
						Events: []syncEvent{
							{Type: "m.room.member", Content: map[string]any{"membership": "invite"}, Sender: "@other:localhost"},
						},
					},
				},
			},
		},
	}

	invites := parseInvites(resp)
	if len(invites) != 2 {
		t.Fatalf("invites = %d, want 2", len(invites))
	}

	byRoom := map[string]Invite{}
	for _, inv := range invites {
		byRoom[inv.RoomID] = inv
	}

	if inv, ok := byRoom["!inv1:localhost"]; !ok {
		t.Error("missing !inv1:localhost")
	} else {
		if inv.Name != "Room A" {
			t.Errorf("inv1 name = %q", inv.Name)
		}
		if inv.Sender != "@admin:localhost" {
			t.Errorf("inv1 sender = %q", inv.Sender)
		}
	}

	if inv, ok := byRoom["!inv2:localhost"]; !ok {
		t.Error("missing !inv2:localhost")
	} else {
		if inv.Name != "" {
			t.Errorf("inv2 name = %q, want empty (no m.room.name event)", inv.Name)
		}
		if inv.Sender != "@other:localhost" {
			t.Errorf("inv2 sender = %q", inv.Sender)
		}
	}
}

func TestParseInvitesEmpty(t *testing.T) {
	resp := &syncResponse{
		Rooms: syncRooms{
			Invite: map[string]syncInvitedRoom{},
		},
	}
	invites := parseInvites(resp)
	if len(invites) != 0 {
		t.Errorf("invites = %d, want 0", len(invites))
	}
}

func TestResolveRoomByID(t *testing.T) {
	svc := &service{homeserver: "https://matrix.localhost"}
	resolved, err := svc.resolveRoom(context.Background(), "!abc:localhost")
	if err != nil {
		t.Fatalf("resolveRoom by ID: %v", err)
	}
	if resolved != "!abc:localhost" {
		t.Errorf("resolved = %q, want !abc:localhost", resolved)
	}
}

func TestServerName(t *testing.T) {
	svc := &service{homeserver: "https://matrix.example.com:8448"}
	if got := svc.serverName(); got != "matrix.example.com" {
		t.Errorf("serverName = %q, want matrix.example.com", got)
	}
}

func TestServerNamePlain(t *testing.T) {
	svc := &service{homeserver: "https://localhost"}
	if got := svc.serverName(); got != "localhost" {
		t.Errorf("serverName = %q, want localhost", got)
	}
}
