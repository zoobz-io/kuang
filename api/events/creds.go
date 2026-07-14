package events

import "github.com/zoobz-io/capitan"

// Credential store signals.
var (
	CredentialResolved = capitan.NewSignal("creds.resolved", "Credential resolved for agent")
	CredentialSet      = capitan.NewSignal("creds.set", "Credential created or updated")
	CredentialDeleted  = capitan.NewSignal("creds.deleted", "Credential deleted")
	CredentialListed   = capitan.NewSignal("creds.listed", "Credential keys listed for agent")
	CredentialNotFound = capitan.NewSignal("creds.not_found", "Credential lookup found no matching entry")
	CredentialFailed   = capitan.NewSignal("creds.failed", "Credential store operation failed")
)

// Credential store field keys. The secret value is intentionally never emitted.
var (
	CredentialAgentKey     = capitan.NewStringKey("creds.agent")
	CredentialKeyNameKey   = capitan.NewStringKey("creds.key")
	CredentialCountKey     = capitan.NewIntKey("creds.count")
	CredentialOperationKey = capitan.NewStringKey("creds.operation")
	CredentialErrorKey     = capitan.NewStringKey("creds.error")
)
