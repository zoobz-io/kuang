package github

import "testing"

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name:    "owner and api url set",
			cfg:     Config{Owner: "octocat", APIURL: "https://api.github.com"},
			wantErr: false,
		},
		{
			name:    "enterprise api url",
			cfg:     Config{Owner: "octocat", APIURL: "https://ghe.example.com/api/v3"},
			wantErr: false,
		},
		{
			name:    "owner required",
			cfg:     Config{APIURL: "https://api.github.com"},
			wantErr: true,
		},
		{
			name:    "api url required",
			cfg:     Config{Owner: "octocat"},
			wantErr: true,
		},
		{
			name:    "api url must be valid url",
			cfg:     Config{Owner: "octocat", APIURL: "not a url"},
			wantErr: true,
		},
		{
			name:    "empty",
			cfg:     Config{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
