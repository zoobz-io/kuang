package wire

import "testing"

func TestCredentialResponseValidate(t *testing.T) {
	tests := []struct {
		name    string
		resp    CredentialResponse
		wantErr bool
	}{
		{"valid", CredentialResponse{Key: "api-token", Value: "secret"}, false},
		{"missing key", CredentialResponse{Value: "secret"}, true},
		{"missing value", CredentialResponse{Key: "api-token"}, true},
		{"missing both", CredentialResponse{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.resp.Validate()
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected valid, got %v", err)
			}
		})
	}
}

func TestSetCredentialRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		req     SetCredentialRequest
		wantErr bool
	}{
		{"valid", SetCredentialRequest{Value: "secret"}, false},
		{"missing value", SetCredentialRequest{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected valid, got %v", err)
			}
		})
	}
}

// CredentialKeysResponse has no Validate method (unlike the other wire types);
// this just confirms it holds and round-trips its keys.
func TestCredentialKeysResponse(t *testing.T) {
	resp := CredentialKeysResponse{Keys: []string{"token-1", "token-2"}}
	if len(resp.Keys) != 2 {
		t.Fatalf("got %d keys, want 2", len(resp.Keys))
	}
	empty := CredentialKeysResponse{}
	if len(empty.Keys) != 0 {
		t.Errorf("zero-value should have no keys, got %d", len(empty.Keys))
	}
}

func TestCredentialRefValidate(t *testing.T) {
	tests := []struct {
		name    string
		ref     CredentialRef
		wantErr bool
	}{
		{"valid", CredentialRef{Key: "api-token"}, false},
		{"missing key", CredentialRef{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ref.Validate()
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected valid, got %v", err)
			}
		})
	}
}
