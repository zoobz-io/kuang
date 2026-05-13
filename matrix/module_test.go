package matrix

import "testing"

func TestModuleFactoryValidConfig(t *testing.T) {
	cfg := Config{
		Homeserver:  "https://matrix.example.com",
		AccessToken: "syt_test_token",
	}
	mod := Module(cfg)
	if mod == nil {
		t.Fatal("expected non-nil module")
	}
}

func TestModuleFactoryEmptyConfig(t *testing.T) {
	mod := Module(Config{})
	if mod == nil {
		t.Fatal("expected non-nil module (validation happens on call)")
	}
}
