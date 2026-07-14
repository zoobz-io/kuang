package matrix

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/zoobzio/kuang/internal/httpc"
	"github.com/zoobzio/kuang/modules/matrix/models"
)

const msgTypeRoomMessage = "m.room.message"

// API implements Matrix API operations via the instrumented HTTP client.
type API struct {
	client     *httpc.Client
	homeserver string
}

// NewAPI constructs an API service from the given configuration.
func NewAPI(cfg Config) *API {
	hs := strings.TrimRight(cfg.Homeserver, "/")
	client := httpc.New(
		httpc.WithBaseURL(hs),
	)
	return &API{client: client, homeserver: hs}
}

// serverName extracts the hostname from the homeserver URL.
func (s *API) serverName() string {
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

// WhoAmI returns the identity of the authenticated user.
func (s *API) WhoAmI(ctx context.Context, opts ...httpc.RequestOption) (models.Identity, error) {
	resp, err := s.client.Get(ctx, "/_matrix/client/v3/account/whoami", opts...)
	if err != nil {
		return models.Identity{}, err
	}
	var id models.Identity
	if err := resp.Decode(&id); err != nil {
		return models.Identity{}, err
	}
	return id, nil
}

// --- Rooms ---

// CreateRoom creates a room. A non-empty alias makes it a public room.
func (s *API) CreateRoom(ctx context.Context, name, topic, alias string, opts ...httpc.RequestOption) (models.Room, error) {
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
		return models.Room{}, err
	}
	var room models.Room
	if err := resp.Decode(&room); err != nil {
		return models.Room{}, err
	}
	return room, nil
}

// JoinedRooms returns the rooms the user has joined, enriched with name and topic.
func (s *API) JoinedRooms(ctx context.Context, opts ...httpc.RequestOption) (models.RoomList, error) {
	resp, err := s.client.Get(ctx, "/_matrix/client/v3/joined_rooms", opts...)
	if err != nil {
		return models.RoomList{}, err
	}
	var joined struct {
		Rooms []string `json:"joined_rooms"`
	}
	if err := resp.Decode(&joined); err != nil {
		return models.RoomList{}, err
	}
	rooms := make([]models.RoomInfo, 0, len(joined.Rooms))
	for _, roomID := range joined.Rooms {
		info := models.RoomInfo{RoomID: roomID}
		if ri, err := s.getRoomInfo(ctx, roomID, opts...); err == nil {
			info.Name = ri.Name
			info.Topic = ri.Topic
		}
		rooms = append(rooms, info)
	}
	return models.RoomList{Rooms: rooms}, nil
}

func (s *API) getRoomInfo(ctx context.Context, roomID string, opts ...httpc.RequestOption) (models.RoomInfo, error) {
	var info models.RoomInfo
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

// Join joins a room by ID or alias, resolving the alias to an ID if needed.
func (s *API) Join(ctx context.Context, roomIDOrAlias string, opts ...httpc.RequestOption) (models.Room, error) {
	resolved, err := s.resolveRoom(ctx, roomIDOrAlias, opts...)
	if err != nil {
		return models.Room{}, err
	}
	resp, err := s.client.Post(ctx, fmt.Sprintf("/_matrix/client/v3/join/%s", escapePathParam(resolved)), map[string]any{}, opts...)
	if err != nil {
		return models.Room{}, err
	}
	var room models.Room
	if err := resp.Decode(&room); err != nil {
		return models.Room{}, err
	}
	return room, nil
}

// Leave leaves a room.
func (s *API) Leave(ctx context.Context, roomID string, opts ...httpc.RequestOption) error {
	_, err := s.client.Post(ctx, fmt.Sprintf("/_matrix/client/v3/rooms/%s/leave", escapePathParam(roomID)), map[string]any{}, opts...)
	return err
}

func (s *API) resolveRoom(ctx context.Context, roomIDOrAlias string, opts ...httpc.RequestOption) (string, error) {
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

// Members returns the joined members of a room.
func (s *API) Members(ctx context.Context, roomID string, opts ...httpc.RequestOption) (models.MemberList, error) {
	resp, err := s.client.Get(ctx, fmt.Sprintf("/_matrix/client/v3/rooms/%s/joined_members", escapePathParam(roomID)), opts...)
	if err != nil {
		return models.MemberList{}, err
	}
	var raw struct {
		Joined map[string]struct {
			DisplayName string `json:"display_name"`
		} `json:"joined"`
	}
	if err := resp.Decode(&raw); err != nil {
		return models.MemberList{}, err
	}
	members := make([]models.Member, 0, len(raw.Joined))
	for userID, info := range raw.Joined {
		members = append(members, models.Member{
			UserID:      userID,
			DisplayName: info.DisplayName,
		})
	}
	return models.MemberList{Members: members}, nil
}

// --- Messages ---

// Send posts a text message to a room.
func (s *API) Send(ctx context.Context, roomID, message string, opts ...httpc.RequestOption) (models.SendMessageResponse, error) {
	txnID := fmt.Sprintf("%d", time.Now().UnixNano())
	body := map[string]any{
		"msgtype": "m.text",
		"body":    message,
	}
	path := fmt.Sprintf("/_matrix/client/v3/rooms/%s/send/m.room.message/%s", escapePathParam(roomID), txnID)
	resp, err := s.client.Put(ctx, path, body, opts...)
	if err != nil {
		return models.SendMessageResponse{}, err
	}
	var result models.SendMessageResponse
	if err := resp.Decode(&result); err != nil {
		return models.SendMessageResponse{}, err
	}
	return result, nil
}

// ReadMessages returns the most recent messages in a room, in chronological
// order, optionally filtered by sender.
func (s *API) ReadMessages(ctx context.Context, roomID string, limit int, from string, opts ...httpc.RequestOption) (models.MessageList, error) {
	path := fmt.Sprintf("/_matrix/client/v3/rooms/%s/messages?dir=b&limit=%d", escapePathParam(roomID), limit)
	return s.decodeMessages(ctx, path, from, true, opts...)
}

// ReadMessagesSince retrieves messages after a given event ID. This is used
// by agents to catch up on messages they haven't seen yet.
func (s *API) ReadMessagesSince(ctx context.Context, roomID, eventID string, limit int, from string, opts ...httpc.RequestOption) (models.MessageList, error) {
	// Resolve the event ID to a pagination token.
	ctxPath := fmt.Sprintf("/_matrix/client/v3/rooms/%s/context/%s?limit=0",
		escapePathParam(roomID), escapePathParam(eventID))
	resp, err := s.client.Get(ctx, ctxPath, opts...)
	if err != nil {
		return models.MessageList{}, fmt.Errorf("resolving event %s: %w", eventID, err)
	}
	var ctxResp struct {
		End string `json:"end"`
	}
	if err := resp.Decode(&ctxResp); err != nil {
		return models.MessageList{}, err
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

func (s *API) decodeMessages(ctx context.Context, path, from string, reverse bool, opts ...httpc.RequestOption) (models.MessageList, error) {
	resp, err := s.client.Get(ctx, path, opts...)
	if err != nil {
		return models.MessageList{}, err
	}
	var raw struct {
		Chunk []rawEvent `json:"chunk"`
	}
	if err := resp.Decode(&raw); err != nil {
		return models.MessageList{}, err
	}
	msgs := make([]models.Message, 0, len(raw.Chunk))
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
	return models.MessageList{Messages: msgs}, nil
}

type rawEvent struct {
	Sender  string         `json:"sender"`
	Type    string         `json:"type"`
	Content map[string]any `json:"content"`
	EventID string         `json:"event_id"`
}

func extractMessage(e rawEvent, from string) (models.Message, bool) {
	if e.Type != msgTypeRoomMessage {
		return models.Message{}, false
	}
	if from != "" && !senderMatches(e.Sender, from) {
		return models.Message{}, false
	}
	body, _ := e.Content["body"].(string)
	return models.Message{
		Sender:  e.Sender,
		Body:    body,
		EventID: e.EventID,
	}, true
}

// --- Invites ---

// Invite invites a user to a room.
func (s *API) Invite(ctx context.Context, roomID, userID string, opts ...httpc.RequestOption) error {
	_, err := s.client.Post(ctx, fmt.Sprintf("/_matrix/client/v3/rooms/%s/invite", escapePathParam(roomID)), map[string]any{
		"user_id": userID,
	}, opts...)
	return err
}

// ListInvites returns the rooms the user has been invited to.
func (s *API) ListInvites(ctx context.Context, opts ...httpc.RequestOption) (models.InviteList, error) {
	resp, err := s.sync(ctx, "", 0, "", opts...)
	if err != nil {
		return models.InviteList{}, err
	}
	return models.InviteList{Invites: parseInvites(resp)}, nil
}

// --- DMs ---

// DMSend sends a direct message to a user, creating the DM room if needed.
func (s *API) DMSend(ctx context.Context, userID, message string, opts ...httpc.RequestOption) (models.SendMessageResponse, error) {
	roomID, err := s.ensureDMRoom(ctx, userID, opts...)
	if err != nil {
		return models.SendMessageResponse{}, err
	}
	return s.Send(ctx, roomID, message, opts...)
}

// DMRead returns the most recent direct messages with a user.
func (s *API) DMRead(ctx context.Context, userID string, limit int, opts ...httpc.RequestOption) (models.MessageList, error) {
	roomID, err := s.findDMRoom(ctx, userID, opts...)
	if err != nil {
		return models.MessageList{}, err
	}
	return s.ReadMessages(ctx, roomID, limit, "", opts...)
}

func (s *API) ensureDMRoom(ctx context.Context, targetID string, opts ...httpc.RequestOption) (string, error) {
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

func (s *API) findDMRoom(ctx context.Context, targetID string, opts ...httpc.RequestOption) (string, error) {
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

func (s *API) createDMRoom(ctx context.Context, invite string, opts ...httpc.RequestOption) (models.Room, error) {
	resp, err := s.client.Post(ctx, "/_matrix/client/v3/createRoom", map[string]any{
		"preset":    "trusted_private_chat",
		"is_direct": true,
		"invite":    []string{invite},
	}, opts...)
	if err != nil {
		return models.Room{}, err
	}
	var room models.Room
	if err := resp.Decode(&room); err != nil {
		return models.Room{}, err
	}
	return room, nil
}

func (s *API) getDirectRooms(ctx context.Context, userID string, opts ...httpc.RequestOption) (map[string][]string, error) {
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

func (s *API) setDirectRooms(ctx context.Context, userID string, rooms map[string][]string, opts ...httpc.RequestOption) error {
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
func (s *API) Watch(ctx context.Context, since string, timeout int, opts ...httpc.RequestOption) (models.WatchResponse, error) {
	// Initial sync to establish position if no since token.
	if since == "" {
		resp, err := s.sync(ctx, "", 0, "", opts...)
		if err != nil {
			return models.WatchResponse{}, fmt.Errorf("initial sync: %w", err)
		}
		since = resp.NextBatch
	}

	// Long-poll for new events.
	resp, err := s.sync(ctx, since, timeout, "", opts...)
	if err != nil {
		return models.WatchResponse{}, err
	}

	events := make([]models.WatchEvent, 0, len(resp.Rooms.Join))

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
			events = append(events, models.WatchEvent{
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
		events = append(events, models.WatchEvent{
			Type:     "invite",
			RoomID:   inv.RoomID,
			RoomName: name,
			Sender:   sender,
			Body:     fmt.Sprintf("invited you to %s", name),
		})
	}

	return models.WatchResponse{
		Events:    events,
		NextBatch: resp.NextBatch,
	}, nil
}

// sync performs a raw Matrix sync request.
func (s *API) sync(ctx context.Context, since string, timeout int, roomID string, opts ...httpc.RequestOption) (*syncResponse, error) {
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
func parseInvites(resp *syncResponse) []models.Invite {
	invites := make([]models.Invite, 0, len(resp.Rooms.Invite))
	for roomID, room := range resp.Rooms.Invite {
		inv := models.Invite{RoomID: roomID}
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
