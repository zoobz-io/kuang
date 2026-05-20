package matrix

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/zoobz-io/kuang/httpc"
)

const msgTypeRoomMessage = "m.room.message"

// service implements Matrix API operations via the instrumented HTTP client.
type service struct {
	client     *httpc.Client
	homeserver string
}

func newService(cfg Config) *service {
	hs := strings.TrimRight(cfg.Homeserver, "/")
	client := httpc.New(
		httpc.WithBaseURL(hs),
	)
	return &service{client: client, homeserver: hs}
}

// serverName extracts the hostname from the homeserver URL.
func (s *service) serverName() string {
	u, err := url.Parse(s.homeserver)
	if err != nil {
		return s.homeserver
	}
	host := u.Hostname()
	if host == "" {
		return s.homeserver
	}
	return host
}

// escapePathParam percent-encodes a value for use in a URL path segment.
func escapePathParam(v string) string {
	return strings.ReplaceAll(url.PathEscape(v), ":", "%3A")
}

// --- Identity ---

func (s *service) WhoAmI(ctx context.Context, opts ...httpc.RequestOption) (Identity, error) {
	resp, err := s.client.Get(ctx, "/_matrix/client/v3/account/whoami", opts...)
	if err != nil {
		return Identity{}, err
	}
	var id Identity
	if err := resp.Decode(&id); err != nil {
		return Identity{}, err
	}
	return id, nil
}

// --- Rooms ---

func (s *service) CreateRoom(ctx context.Context, name, topic, alias string, opts ...httpc.RequestOption) (Room, error) {
	body := map[string]any{
		"name":   name,
		"preset": "private_chat",
	}
	if topic != "" {
		body["topic"] = topic
	}
	if alias != "" {
		body["preset"] = "public_chat"
		body["room_alias_name"] = alias
	}
	resp, err := s.client.Post(ctx, "/_matrix/client/v3/createRoom", body, opts...)
	if err != nil {
		return Room{}, err
	}
	var room Room
	if err := resp.Decode(&room); err != nil {
		return Room{}, err
	}
	return room, nil
}

func (s *service) JoinedRooms(ctx context.Context, opts ...httpc.RequestOption) (RoomList, error) {
	resp, err := s.client.Get(ctx, "/_matrix/client/v3/joined_rooms", opts...)
	if err != nil {
		return RoomList{}, err
	}
	var joined struct {
		Rooms []string `json:"joined_rooms"`
	}
	if err := resp.Decode(&joined); err != nil {
		return RoomList{}, err
	}
	rooms := make([]RoomInfo, 0, len(joined.Rooms))
	for _, roomID := range joined.Rooms {
		info := RoomInfo{RoomID: roomID}
		if ri, err := s.getRoomInfo(ctx, roomID, opts...); err == nil {
			info.Name = ri.Name
			info.Topic = ri.Topic
		}
		rooms = append(rooms, info)
	}
	return RoomList{Rooms: rooms}, nil
}

func (s *service) getRoomInfo(ctx context.Context, roomID string, opts ...httpc.RequestOption) (RoomInfo, error) {
	var info RoomInfo
	var nameEvent struct {
		Name string `json:"name"`
	}
	if resp, err := s.client.Get(ctx, fmt.Sprintf("/_matrix/client/v3/rooms/%s/state/m.room.name", escapePathParam(roomID)), opts...); err == nil {
		if err := resp.Decode(&nameEvent); err == nil {
			info.Name = nameEvent.Name
		}
	}
	var topicEvent struct {
		Topic string `json:"topic"`
	}
	if resp, err := s.client.Get(ctx, fmt.Sprintf("/_matrix/client/v3/rooms/%s/state/m.room.topic", escapePathParam(roomID)), opts...); err == nil {
		if err := resp.Decode(&topicEvent); err == nil {
			info.Topic = topicEvent.Topic
		}
	}
	return info, nil
}

func (s *service) Join(ctx context.Context, roomIDOrAlias string, opts ...httpc.RequestOption) (Room, error) {
	resolved, err := s.resolveRoom(ctx, roomIDOrAlias, opts...)
	if err != nil {
		return Room{}, err
	}
	resp, err := s.client.Post(ctx, fmt.Sprintf("/_matrix/client/v3/join/%s", escapePathParam(resolved)), map[string]any{}, opts...)
	if err != nil {
		return Room{}, err
	}
	var room Room
	if err := resp.Decode(&room); err != nil {
		return Room{}, err
	}
	return room, nil
}

func (s *service) Leave(ctx context.Context, roomID string, opts ...httpc.RequestOption) error {
	_, err := s.client.Post(ctx, fmt.Sprintf("/_matrix/client/v3/rooms/%s/leave", escapePathParam(roomID)), map[string]any{}, opts...)
	return err
}

func (s *service) resolveRoom(ctx context.Context, roomIDOrAlias string, opts ...httpc.RequestOption) (string, error) {
	if strings.HasPrefix(roomIDOrAlias, "!") {
		return roomIDOrAlias, nil
	}
	alias := roomIDOrAlias
	if !strings.HasPrefix(alias, "#") {
		alias = "#" + alias + ":" + s.serverName()
	}
	resp, err := s.client.Get(ctx, fmt.Sprintf("/_matrix/client/v3/directory/room/%s", escapePathParam(alias)), opts...)
	if err != nil {
		return "", fmt.Errorf("resolving %q: %w", alias, err)
	}
	var ar struct {
		RoomID string `json:"room_id"`
	}
	if err := resp.Decode(&ar); err != nil {
		return "", err
	}
	return ar.RoomID, nil
}

// --- Members ---

func (s *service) Members(ctx context.Context, roomID string, opts ...httpc.RequestOption) (MemberList, error) {
	resp, err := s.client.Get(ctx, fmt.Sprintf("/_matrix/client/v3/rooms/%s/joined_members", escapePathParam(roomID)), opts...)
	if err != nil {
		return MemberList{}, err
	}
	var raw struct {
		Joined map[string]struct {
			DisplayName string `json:"display_name"`
		} `json:"joined"`
	}
	if err := resp.Decode(&raw); err != nil {
		return MemberList{}, err
	}
	members := make([]Member, 0, len(raw.Joined))
	for userID, info := range raw.Joined {
		members = append(members, Member{
			UserID:      userID,
			DisplayName: info.DisplayName,
		})
	}
	return MemberList{Members: members}, nil
}

// --- Messages ---

func (s *service) Send(ctx context.Context, roomID, message string, opts ...httpc.RequestOption) (SendMessageResponse, error) {
	txnID := fmt.Sprintf("%d", time.Now().UnixNano())
	body := map[string]any{
		"msgtype": "m.text",
		"body":    message,
	}
	path := fmt.Sprintf("/_matrix/client/v3/rooms/%s/send/m.room.message/%s", escapePathParam(roomID), txnID)
	resp, err := s.client.Put(ctx, path, body, opts...)
	if err != nil {
		return SendMessageResponse{}, err
	}
	var result SendMessageResponse
	if err := resp.Decode(&result); err != nil {
		return SendMessageResponse{}, err
	}
	return result, nil
}

func (s *service) ReadMessages(ctx context.Context, roomID string, limit int, from string, opts ...httpc.RequestOption) (MessageList, error) {
	path := fmt.Sprintf("/_matrix/client/v3/rooms/%s/messages?dir=b&limit=%d", escapePathParam(roomID), limit)
	return s.decodeMessages(ctx, path, from, true, opts...)
}

// ReadMessagesSince retrieves messages after a given event ID. This is used
// by agents to catch up on messages they haven't seen yet.
func (s *service) ReadMessagesSince(ctx context.Context, roomID, eventID string, limit int, from string, opts ...httpc.RequestOption) (MessageList, error) {
	// Resolve the event ID to a pagination token.
	ctxPath := fmt.Sprintf("/_matrix/client/v3/rooms/%s/context/%s?limit=0",
		escapePathParam(roomID), escapePathParam(eventID))
	resp, err := s.client.Get(ctx, ctxPath, opts...)
	if err != nil {
		return MessageList{}, fmt.Errorf("resolving event %s: %w", eventID, err)
	}
	var ctxResp struct {
		End string `json:"end"`
	}
	if err := resp.Decode(&ctxResp); err != nil {
		return MessageList{}, err
	}
	// Read forward from that point (chronological order).
	path := fmt.Sprintf("/_matrix/client/v3/rooms/%s/messages?dir=f&from=%s&limit=%d",
		escapePathParam(roomID), url.QueryEscape(ctxResp.End), limit)
	return s.decodeMessages(ctx, path, from, false, opts...)
}

// senderMatches checks if a message sender contains the filter string.
func senderMatches(sender, filter string) bool {
	return strings.Contains(sender, filter)
}

func (s *service) decodeMessages(ctx context.Context, path, from string, reverse bool, opts ...httpc.RequestOption) (MessageList, error) {
	resp, err := s.client.Get(ctx, path, opts...)
	if err != nil {
		return MessageList{}, err
	}
	var raw struct {
		Chunk []struct {
			Sender  string         `json:"sender"`
			Type    string         `json:"type"`
			Content map[string]any `json:"content"`
			EventID string         `json:"event_id"`
		} `json:"chunk"`
	}
	if err := resp.Decode(&raw); err != nil {
		return MessageList{}, err
	}
	msgs := make([]Message, 0, len(raw.Chunk))
	if reverse {
		for i := len(raw.Chunk) - 1; i >= 0; i-- {
			if m, ok := extractMessage(raw.Chunk[i], from); ok {
				msgs = append(msgs, m)
			}
		}
	} else {
		for _, c := range raw.Chunk {
			if m, ok := extractMessage(c, from); ok {
				msgs = append(msgs, m)
			}
		}
	}
	return MessageList{Messages: msgs}, nil
}

type rawEvent struct {
	Sender  string         `json:"sender"`
	Type    string         `json:"type"`
	Content map[string]any `json:"content"`
	EventID string         `json:"event_id"`
}

func extractMessage(e rawEvent, from string) (Message, bool) {
	if e.Type != msgTypeRoomMessage {
		return Message{}, false
	}
	if from != "" && !senderMatches(e.Sender, from) {
		return Message{}, false
	}
	body, _ := e.Content["body"].(string)
	return Message{
		Sender:  e.Sender,
		Body:    body,
		EventID: e.EventID,
	}, true
}

// --- Invites ---

func (s *service) Invite(ctx context.Context, roomID, userID string, opts ...httpc.RequestOption) error {
	_, err := s.client.Post(ctx, fmt.Sprintf("/_matrix/client/v3/rooms/%s/invite", escapePathParam(roomID)), map[string]any{
		"user_id": userID,
	}, opts...)
	return err
}

func (s *service) ListInvites(ctx context.Context, opts ...httpc.RequestOption) (InviteList, error) {
	resp, err := s.sync(ctx, "", 0, "", opts...)
	if err != nil {
		return InviteList{}, err
	}
	return InviteList{Invites: parseInvites(resp)}, nil
}

// --- DMs ---

func (s *service) DMSend(ctx context.Context, userID, message string, opts ...httpc.RequestOption) (SendMessageResponse, error) {
	roomID, err := s.ensureDMRoom(ctx, userID, opts...)
	if err != nil {
		return SendMessageResponse{}, err
	}
	return s.Send(ctx, roomID, message, opts...)
}

func (s *service) DMRead(ctx context.Context, userID string, limit int, opts ...httpc.RequestOption) (MessageList, error) {
	roomID, err := s.findDMRoom(ctx, userID, opts...)
	if err != nil {
		return MessageList{}, err
	}
	return s.ReadMessages(ctx, roomID, limit, "", opts...)
}

func (s *service) ensureDMRoom(ctx context.Context, targetID string, opts ...httpc.RequestOption) (string, error) {
	me, err := s.WhoAmI(ctx, opts...)
	if err != nil {
		return "", fmt.Errorf("getting identity: %w", err)
	}
	directs, err := s.getDirectRooms(ctx, me.UserID, opts...)
	if err != nil {
		return "", err
	}
	if rooms, ok := directs[targetID]; ok && len(rooms) > 0 {
		return rooms[0], nil
	}
	room, err := s.createDMRoom(ctx, targetID, opts...)
	if err != nil {
		return "", err
	}
	directs[targetID] = []string{room.RoomID}
	if err := s.setDirectRooms(ctx, me.UserID, directs, opts...); err != nil {
		return "", err
	}
	return room.RoomID, nil
}

func (s *service) findDMRoom(ctx context.Context, targetID string, opts ...httpc.RequestOption) (string, error) {
	me, err := s.WhoAmI(ctx, opts...)
	if err != nil {
		return "", fmt.Errorf("getting identity: %w", err)
	}
	directs, err := s.getDirectRooms(ctx, me.UserID, opts...)
	if err != nil {
		return "", err
	}
	rooms, ok := directs[targetID]
	if !ok || len(rooms) == 0 {
		return "", fmt.Errorf("no DM room with %s", targetID)
	}
	return rooms[0], nil
}

func (s *service) createDMRoom(ctx context.Context, invite string, opts ...httpc.RequestOption) (Room, error) {
	resp, err := s.client.Post(ctx, "/_matrix/client/v3/createRoom", map[string]any{
		"preset":    "trusted_private_chat",
		"is_direct": true,
		"invite":    []string{invite},
	}, opts...)
	if err != nil {
		return Room{}, err
	}
	var room Room
	if err := resp.Decode(&room); err != nil {
		return Room{}, err
	}
	return room, nil
}

func (s *service) getDirectRooms(ctx context.Context, userID string, opts ...httpc.RequestOption) (map[string][]string, error) {
	path := fmt.Sprintf("/_matrix/client/v3/user/%s/account_data/m.direct", escapePathParam(userID))
	resp, err := s.client.Get(ctx, path, opts...)
	if err != nil {
		return map[string][]string{}, nil
	}
	var directs map[string][]string
	if err := resp.Decode(&directs); err != nil {
		return map[string][]string{}, nil
	}
	return directs, nil
}

func (s *service) setDirectRooms(ctx context.Context, userID string, rooms map[string][]string, opts ...httpc.RequestOption) error {
	path := fmt.Sprintf("/_matrix/client/v3/user/%s/account_data/m.direct", escapePathParam(userID))
	_, err := s.client.Put(ctx, path, rooms, opts...)
	return err
}

// --- Watch (long-poll via /sync) ---

// Watch performs a long-poll against the Matrix /sync endpoint. It blocks
// until new messages arrive or the timeout expires. The since parameter is
// the next_batch token from a previous sync — pass empty string for initial
// sync. The timeout controls how long the Matrix server holds the connection
// open waiting for events (in seconds).
func (s *service) Watch(ctx context.Context, since string, timeout int, opts ...httpc.RequestOption) (WatchResponse, error) {
	// Initial sync to establish position if no since token.
	if since == "" {
		resp, err := s.sync(ctx, "", 0, "", opts...)
		if err != nil {
			return WatchResponse{}, fmt.Errorf("initial sync: %w", err)
		}
		since = resp.NextBatch
	}

	// Long-poll for new events.
	resp, err := s.sync(ctx, since, timeout, "", opts...)
	if err != nil {
		return WatchResponse{}, err
	}

	var events []WatchEvent

	// Collect messages from joined rooms.
	for roomID, room := range resp.Rooms.Join {
		roomName := roomID
		if info, infoErr := s.getRoomInfo(ctx, roomID, opts...); infoErr == nil && info.Name != "" {
			roomName = info.Name
		}
		for _, ev := range room.Timeline.Events {
			if ev.Type != msgTypeRoomMessage {
				continue
			}
			body, _ := ev.Content["body"].(string)
			events = append(events, WatchEvent{
				Type:     "message",
				RoomID:   roomID,
				RoomName: roomName,
				Sender:   ev.Sender,
				Body:     body,
				EventID:  ev.EventID,
			})
		}
	}

	// Collect invites.
	for _, inv := range parseInvites(resp) {
		sender := inv.Sender
		if sender == "" {
			sender = "unknown"
		}
		name := inv.Name
		if name == "" {
			name = inv.RoomID
		}
		events = append(events, WatchEvent{
			Type:     "invite",
			RoomID:   inv.RoomID,
			RoomName: name,
			Sender:   sender,
			Body:     fmt.Sprintf("invited you to %s", name),
		})
	}

	return WatchResponse{
		Events:    events,
		NextBatch: resp.NextBatch,
	}, nil
}

// sync performs a raw Matrix sync request.
func (s *service) sync(ctx context.Context, since string, timeout int, roomID string, opts ...httpc.RequestOption) (*syncResponse, error) {
	q := url.Values{}
	q.Set("timeout", fmt.Sprintf("%d", timeout*1000))
	if since != "" {
		q.Set("since", since)
	}
	if roomID != "" {
		filter := fmt.Sprintf(`{"room":{"rooms":[%q],"timeline":{"limit":10}}}`, roomID)
		q.Set("filter", filter)
	}
	path := "/_matrix/client/v3/sync?" + q.Encode()
	resp, err := s.client.Get(ctx, path, opts...)
	if err != nil {
		return nil, err
	}
	var sr syncResponse
	if err := resp.Decode(&sr); err != nil {
		return nil, err
	}
	return &sr, nil
}

// --- internal sync types ---

type syncResponse struct {
	Rooms     syncRooms `json:"rooms"`
	NextBatch string    `json:"next_batch"`
}

type syncRooms struct {
	Join   map[string]syncJoinedRoom  `json:"join"`
	Invite map[string]syncInvitedRoom `json:"invite"`
}

type syncJoinedRoom struct {
	Timeline syncTimeline `json:"timeline"`
}

type syncInvitedRoom struct {
	InviteState syncInviteState `json:"invite_state"`
}

type syncTimeline struct {
	Events []syncEvent `json:"events"`
}

type syncInviteState struct {
	Events []syncEvent `json:"events"`
}

type syncEvent struct {
	Sender  string         `json:"sender"`
	Type    string         `json:"type"`
	Content map[string]any `json:"content"`
	EventID string         `json:"event_id"`
}

// parseInvites extracts invite info from a sync response.
func parseInvites(resp *syncResponse) []Invite {
	invites := make([]Invite, 0, len(resp.Rooms.Invite))
	for roomID, room := range resp.Rooms.Invite {
		inv := Invite{RoomID: roomID}
		for _, ev := range room.InviteState.Events {
			switch ev.Type {
			case "m.room.name":
				if name, ok := ev.Content["name"].(string); ok {
					inv.Name = name
				}
			case "m.room.member":
				if membership, ok := ev.Content["membership"].(string); ok && membership == "invite" {
					inv.Sender = ev.Sender
				}
			}
		}
		invites = append(invites, inv)
	}
	return invites
}
