// Package events defines capitan signals for kuang lifecycle events.
package events

import "github.com/zoobz-io/capitan"

// Startup lifecycle signals.
var (
	StartupDatabaseConnected = capitan.NewSignal("startup.database.connected", "Database connection established")
	StartupStorageConnected  = capitan.NewSignal("startup.storage.connected", "Storage connection established")
	StartupServicesReady     = capitan.NewSignal("startup.services.ready", "All services registered and frozen")
	StartupOTELReady         = capitan.NewSignal("startup.otel.ready", "OpenTelemetry providers initialized")
	StartupApertureReady     = capitan.NewSignal("startup.aperture.ready", "Aperture bridge initialized")
	StartupServerListening   = capitan.NewSignal("startup.server.listening", "HTTP server accepting connections")
	StartupFailed            = capitan.NewSignal("startup.failed", "Startup sequence failed")
)

// Startup field keys.
var (
	StartupPortKey    = capitan.NewIntKey("startup.port")
	StartupWorkersKey = capitan.NewIntKey("startup.workers")
	StartupErrorKey   = capitan.NewErrorKey("startup.error")
)
