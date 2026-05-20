package matrix

import (
	"context"
	"strconv"

	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/kuang/contracts"
	"github.com/zoobz-io/kuang/httpc"
	"github.com/zoobz-io/kuang/internal/auth"
)

// CredentialKey is the credential key used to resolve Matrix tokens.
const CredentialKey = "matrix-token"

// Scope constants for the Matrix module. These are used in kuang.yaml to
// grant agents access to specific operations. Patterns like "matrix-*-read"
// match all read scopes via the CLI's glob matching.
const (
	ScopeRoomsRead     = "matrix-rooms-read"
	ScopeRoomsWrite    = "matrix-rooms-write"
	ScopeMessagesRead  = "matrix-messages-read"
	ScopeMessagesWrite = "matrix-messages-write"
	ScopeMembersRead   = "matrix-members-read"
	ScopeInvitesRead   = "matrix-invites-read"
	ScopeInvitesWrite  = "matrix-invites-write"
	ScopeWatchRead     = "matrix-watch-read"
	ScopeDMRead        = "matrix-dm-read"
	ScopeDMWrite       = "matrix-dm-write"
	ScopeIdentityRead  = "matrix-identity-read"
)

// resolveAuth resolves the agent's Matrix bearer token from the credential store.
func resolveAuth(ctx context.Context) (httpc.RequestOption, error) {
	creds, err := sum.Use[contracts.Credentials](ctx)
	if err != nil {
		return nil, rocco.ErrForbidden.WithMessage("credential store unavailable")
	}
	identity := auth.IdentityFromContext(ctx)
	if identity == nil {
		return nil, rocco.ErrUnauthorized.WithMessage("identity required")
	}
	token, err := creds.Resolve(ctx, identity.ID(), CredentialKey)
	if err != nil {
		return nil, rocco.ErrForbidden.WithMessage("matrix credentials not configured")
	}
	return httpc.WithRequestBearerToken(token), nil
}

func endpoints(svc *service) []rocco.Endpoint {
	return []rocco.Endpoint{
		whoAmI(svc),
		listRooms(svc),
		createRoom(svc),
		joinRoom(svc),
		leaveRoom(svc),
		listMembers(svc),
		sendMessage(svc),
		readMessages(svc),
		inviteUser(svc),
		listInvites(svc),
		dmSend(svc),
		dmRead(svc),
		watch(svc),
	}
}

func whoAmI(svc *service) rocco.Endpoint {
	return rocco.GET[rocco.NoBody, Identity]("/matrix/whoami", func(r *rocco.Request[rocco.NoBody]) (Identity, error) {
		auth, err := resolveAuth(r)
		if err != nil {
			return Identity{}, err
		}
		return svc.WhoAmI(r, auth)
	}).WithSummary("Get current identity").WithTags("matrix").WithScopes(ScopeIdentityRead)
}

// --- Rooms ---

func listRooms(svc *service) rocco.Endpoint {
	return rocco.GET[rocco.NoBody, RoomList]("/matrix/rooms", func(r *rocco.Request[rocco.NoBody]) (RoomList, error) {
		auth, err := resolveAuth(r)
		if err != nil {
			return RoomList{}, err
		}
		return svc.JoinedRooms(r, auth)
	}).WithSummary("List joined rooms").WithTags("matrix").WithScopes(ScopeRoomsRead)
}

func createRoom(svc *service) rocco.Endpoint {
	return rocco.POST[CreateRoomRequest, Room]("/matrix/rooms", func(r *rocco.Request[CreateRoomRequest]) (Room, error) {
		auth, err := resolveAuth(r)
		if err != nil {
			return Room{}, err
		}
		return svc.CreateRoom(r, r.Body.Name, r.Body.Topic, r.Body.Alias, auth)
	}).WithSummary("Create a room").WithTags("matrix").WithScopes(ScopeRoomsWrite)
}

func joinRoom(svc *service) rocco.Endpoint {
	return rocco.POST[JoinRoomRequest, Room]("/matrix/rooms/join", func(r *rocco.Request[JoinRoomRequest]) (Room, error) {
		auth, err := resolveAuth(r)
		if err != nil {
			return Room{}, err
		}
		return svc.Join(r, r.Body.RoomIDOrAlias, auth)
	}).WithSummary("Join a room").WithTags("matrix").WithScopes(ScopeRoomsWrite)
}

func leaveRoom(svc *service) rocco.Endpoint {
	return rocco.POST[rocco.NoBody, rocco.NoBody]("/matrix/rooms/{room}/leave", func(r *rocco.Request[rocco.NoBody]) (rocco.NoBody, error) {
		auth, err := resolveAuth(r)
		if err != nil {
			return rocco.NoBody{}, err
		}
		return rocco.NoBody{}, svc.Leave(r, r.Params.Path["room"], auth)
	}).WithSummary("Leave a room").WithTags("matrix").WithPathParams("room").WithScopes(ScopeRoomsWrite)
}

// --- Members ---

func listMembers(svc *service) rocco.Endpoint {
	return rocco.GET[rocco.NoBody, MemberList]("/matrix/rooms/{room}/members", func(r *rocco.Request[rocco.NoBody]) (MemberList, error) {
		auth, err := resolveAuth(r)
		if err != nil {
			return MemberList{}, err
		}
		return svc.Members(r, r.Params.Path["room"], auth)
	}).WithSummary("List room members").WithTags("matrix").WithPathParams("room").WithScopes(ScopeMembersRead)
}

// --- Messages ---

func sendMessage(svc *service) rocco.Endpoint {
	return rocco.POST[SendMessageRequest, SendMessageResponse]("/matrix/rooms/{room}/messages", func(r *rocco.Request[SendMessageRequest]) (SendMessageResponse, error) {
		auth, err := resolveAuth(r)
		if err != nil {
			return SendMessageResponse{}, err
		}
		return svc.Send(r, r.Params.Path["room"], r.Body.Message, auth)
	}).WithSummary("Send a message").WithTags("matrix").WithPathParams("room").WithScopes(ScopeMessagesWrite)
}

func readMessages(svc *service) rocco.Endpoint {
	return rocco.GET[rocco.NoBody, MessageList]("/matrix/rooms/{room}/messages", func(r *rocco.Request[rocco.NoBody]) (MessageList, error) {
		auth, err := resolveAuth(r)
		if err != nil {
			return MessageList{}, err
		}
		limit := 20
		if l := r.Params.Query["limit"]; l != "" {
			if n, err := strconv.Atoi(l); err == nil && n > 0 {
				limit = n
			}
		}
		since := r.Params.Query["since"]
		from := r.Params.Query["from"]
		if since != "" {
			return svc.ReadMessagesSince(r, r.Params.Path["room"], since, limit, from, auth)
		}
		return svc.ReadMessages(r, r.Params.Path["room"], limit, from, auth)
	}).WithSummary("Read messages from a room").WithTags("matrix").WithPathParams("room").WithQueryParams("limit", "since", "from").WithScopes(ScopeMessagesRead)
}

// --- Invites ---

func inviteUser(svc *service) rocco.Endpoint {
	return rocco.POST[InviteRequest, rocco.NoBody]("/matrix/rooms/{room}/invite", func(r *rocco.Request[InviteRequest]) (rocco.NoBody, error) {
		auth, err := resolveAuth(r)
		if err != nil {
			return rocco.NoBody{}, err
		}
		return rocco.NoBody{}, svc.Invite(r, r.Params.Path["room"], r.Body.UserID, auth)
	}).WithSummary("Invite a user to a room").WithTags("matrix").WithPathParams("room").WithScopes(ScopeInvitesWrite)
}

func listInvites(svc *service) rocco.Endpoint {
	return rocco.GET[rocco.NoBody, InviteList]("/matrix/invites", func(r *rocco.Request[rocco.NoBody]) (InviteList, error) {
		auth, err := resolveAuth(r)
		if err != nil {
			return InviteList{}, err
		}
		return svc.ListInvites(r, auth)
	}).WithSummary("List pending invites").WithTags("matrix").WithScopes(ScopeInvitesRead)
}

// --- DMs ---

func dmSend(svc *service) rocco.Endpoint {
	return rocco.POST[DMSendRequest, SendMessageResponse]("/matrix/dm/{user}/messages", func(r *rocco.Request[DMSendRequest]) (SendMessageResponse, error) {
		auth, err := resolveAuth(r)
		if err != nil {
			return SendMessageResponse{}, err
		}
		return svc.DMSend(r, r.Params.Path["user"], r.Body.Message, auth)
	}).WithSummary("Send a direct message").WithTags("matrix").WithPathParams("user").WithScopes(ScopeDMWrite)
}

func dmRead(svc *service) rocco.Endpoint {
	return rocco.GET[rocco.NoBody, MessageList]("/matrix/dm/{user}/messages", func(r *rocco.Request[rocco.NoBody]) (MessageList, error) {
		auth, err := resolveAuth(r)
		if err != nil {
			return MessageList{}, err
		}
		limit := 20
		if l := r.Params.Query["limit"]; l != "" {
			if n, err := strconv.Atoi(l); err == nil && n > 0 {
				limit = n
			}
		}
		return svc.DMRead(r, r.Params.Path["user"], limit, auth)
	}).WithSummary("Read DM history").WithTags("matrix").WithPathParams("user").WithQueryParams("limit").WithScopes(ScopeDMRead)
}

// --- Watch (long-poll) ---

// watch is the core endpoint for agent event loops. It blocks until new
// messages arrive on any joined room or the timeout expires.
func watch(svc *service) rocco.Endpoint {
	return rocco.GET[rocco.NoBody, WatchResponse]("/matrix/watch", func(r *rocco.Request[rocco.NoBody]) (WatchResponse, error) {
		auth, err := resolveAuth(r)
		if err != nil {
			return WatchResponse{}, err
		}
		since := r.Params.Query["since"]
		timeout := 30
		if t := r.Params.Query["timeout"]; t != "" {
			if n, err := strconv.Atoi(t); err == nil && n > 0 {
				timeout = n
			}
		}
		return svc.Watch(r, since, timeout, auth)
	}).WithSummary("Watch for new messages (long-poll)").WithTags("matrix").WithQueryParams("since", "timeout").WithScopes(ScopeWatchRead)
}
