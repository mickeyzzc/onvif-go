package server

import (
	"strings"
	"testing"
	"time"
)

// Config.Validate (issue #63): the whole configuration is checked before
// the server starts — dangerous defaults must fail fast instead of
// surfacing as confusing runtime behavior.

func validConfig() *Config {
	return DefaultConfig()
}

func TestValidateAcceptsDefaultConfig(t *testing.T) {
	if err := DefaultConfig().Validate(); err != nil {
		t.Fatalf("DefaultConfig must validate: %v", err)
	}
}

func TestValidatePortRange(t *testing.T) {
	for _, port := range []int{-1, 65536, 100000} {
		cfg := validConfig()
		cfg.Port = port
		err := cfg.Validate()
		if err == nil || !strings.Contains(err.Error(), "Port") {
			t.Errorf("Port %d: want Port error, got %v", port, err)
		}
	}
	// 0 = kernel-assigned port (legitimate for tests and ephemeral servers).
	for _, port := range []int{0, 1, 80, 8080, 65535} {
		cfg := validConfig()
		cfg.Port = port
		if err := cfg.Validate(); err != nil {
			t.Errorf("Port %d: unexpected error %v", port, err)
		}
	}
}

func TestValidateTimeoutPositive(t *testing.T) {
	for _, timeout := range []time.Duration{0, -time.Second} {
		cfg := validConfig()
		cfg.Timeout = timeout
		err := cfg.Validate()
		if err == nil || !strings.Contains(err.Error(), "Timeout") {
			t.Errorf("Timeout %v: want Timeout error, got %v", timeout, err)
		}
	}
}

func TestValidateHalfConfiguredCredentials(t *testing.T) {
	cfg := validConfig()
	cfg.Username, cfg.Password = "admin", ""
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "credentials") {
		t.Fatalf("username without password: want credentials error, got %v", err)
	}

	cfg = validConfig()
	cfg.Username, cfg.Password = "", "secret"
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "credentials") {
		t.Fatalf("password without username: want credentials error, got %v", err)
	}

	// Both empty (open mode) and both set (auth mode) are valid.
	cfg = validConfig()
	cfg.Username, cfg.Password = "", ""
	if err := cfg.Validate(); err != nil {
		t.Fatalf("open mode: %v", err)
	}
	cfg.Username, cfg.Password = "admin", "secret"
	if err := cfg.Validate(); err != nil {
		t.Fatalf("auth mode: %v", err)
	}
}

func TestValidateAuthProtectedActions(t *testing.T) {
	for _, action := range []string{"", "Set  Scopes", "With\tTab"} {
		cfg := validConfig()
		cfg.Username, cfg.Password = "admin", "secret"
		cfg.AuthProtectedActions = []string{"OkAction", action}
		err := cfg.Validate()
		if err == nil || !strings.Contains(err.Error(), "AuthProtectedActions") {
			t.Errorf("action %q: want AuthProtectedActions error, got %v", action, err)
		}
	}
}

func TestValidateBaseAndSnapshotPaths(t *testing.T) {
	cfg := validConfig()
	cfg.BasePath = "onvif" // missing leading slash
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "BasePath") {
		t.Fatalf("relative BasePath: want BasePath error, got %v", err)
	}

	cfg = validConfig()
	cfg.BasePath = "/onvif/"
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "BasePath") {
		t.Fatalf("trailing-slash BasePath: want BasePath error, got %v", err)
	}

	cfg = validConfig()
	cfg.SnapshotPath = "snap.jpg"
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "SnapshotPath") {
		t.Fatalf("relative SnapshotPath: want SnapshotPath error, got %v", err)
	}
}

func TestValidateProfiles(t *testing.T) {
	cfg := validConfig()
	cfg.Profiles = nil
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "Profiles") {
		t.Fatalf("no profiles: want Profiles error, got %v", err)
	}

	cfg = validConfig()
	cfg.Profiles = []ProfileConfig{
		cfg.Profiles[0],
		cfg.Profiles[0], // duplicate token
	}
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "token") {
		t.Fatalf("duplicate profile tokens: want token error, got %v", err)
	}
}

func TestNewFailsFastOnInvalidConfig(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Port = -1
	if _, err := New(cfg); err == nil {
		t.Fatal("New must reject Port -1 (issue #63 fail-fast)")
	}
}
