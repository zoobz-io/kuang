package httpc

import "github.com/zoobz-io/capitan"

// Signals emitted by the HTTP client.
var (
	RequestStarted   = capitan.NewSignal("httpc.request.started", "External HTTP request initiated")
	RequestCompleted = capitan.NewSignal("httpc.request.completed", "External HTTP request completed successfully")
	RequestError     = capitan.NewSignal("httpc.request.error", "External HTTP request returned error status")
	RequestFailed    = capitan.NewSignal("httpc.request.failed", "External HTTP request failed (network error)")
)

// Field keys for request signals.
var (
	MethodKey     = capitan.NewStringKey("httpc.method")
	URLKey        = capitan.NewStringKey("httpc.url")
	StatusKey     = capitan.NewIntKey("httpc.status")
	DurationMsKey = capitan.NewInt64Key("httpc.duration_ms")
	ErrorKey      = capitan.NewStringKey("httpc.error")
)
