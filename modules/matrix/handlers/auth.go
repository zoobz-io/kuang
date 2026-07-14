package handlers

// CredentialKey is the credential key used to resolve Matrix tokens. Handlers
// pass it to core's authenticated endpoint constructors, which resolve the
// agent's bearer token from the credential store at request time.
const CredentialKey = "matrix-token"
