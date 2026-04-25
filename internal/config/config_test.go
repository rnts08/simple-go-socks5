package config

import (
	"testing"
	"time"
)

func TestConfig_DefaultValues(t *testing.T) {
	cfg := NewConfig(
		":8080", false,
		ModeLocal, "./users.db", "", "",
		ModeLocal, "./traffic.db", "", "",
		60*time.Second, false,
	)

	if cfg.ListenAddr != ":8080" {
		t.Errorf("expected default listen addr :8080, got %s", cfg.ListenAddr)
	}
	if cfg.AuthMode != ModeLocal {
		t.Errorf("expected default auth mode local, got %s", cfg.AuthMode)
	}
	if cfg.AccountingMode != ModeLocal {
		t.Errorf("expected default accounting mode local, got %s", cfg.AccountingMode)
	}
	if cfg.Verbose != false {
		t.Errorf("expected default verbose false, got %v", cfg.Verbose)
	}
	if cfg.AccountingInterval != 60*time.Second {
		t.Errorf("expected default accounting interval 60s, got %v", cfg.AccountingInterval)
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		cfg      *Config
		wantError bool
	}{
		{
			name: "valid local auth",
			cfg: NewConfig(
				":8080", false,
				ModeLocal, "/tmp/test.db", "", "",
				ModeLocal, "/tmp/traffic.db", "", "",
				60*time.Second, false,
			),
			wantError: false,
		},
		{
			name: "valid remote auth",
			cfg: NewConfig(
				":8080", false,
				ModeRemote, "", "https://api.test.com", "",
				ModeLocal, "/tmp/traffic.db", "", "",
				60*time.Second, false,
			),
			wantError: false,
		},
		{
			name: "remote auth without url",
			cfg: NewConfig(
				":8080", false,
				ModeRemote, "", "", "",
				ModeLocal, "/tmp/traffic.db", "", "",
				60*time.Second, false,
			),
			wantError: true,
		},
		{
			name: "valid local accounting",
			cfg: NewConfig(
				":8080", false,
				ModeLocal, "/tmp/users.db", "", "",
				ModeLocal, "/tmp/traffic.db", "", "",
				60*time.Second, false,
			),
			wantError: false,
		},
		{
			name: "valid remote accounting",
			cfg: NewConfig(
				":8080", false,
				ModeLocal, "/tmp/users.db", "", "",
				ModeRemote, "", "https://api.test.com", "",
				60*time.Second, false,
			),
			wantError: false,
		},
		{
			name: "remote accounting without url",
			cfg: NewConfig(
				":8080", false,
				ModeLocal, "/tmp/users.db", "", "",
				ModeRemote, "", "", "",
				60*time.Second, false,
			),
			wantError: true,
		},
		{
			name: "mock mode disables validation",
			cfg: NewConfig(
				":8080", false,
				ModeMock, "", "", "",
				ModeMock, "", "", "",
				60*time.Second, true,
			),
			wantError: false,
		},
		{
			name: "valid both mode",
			cfg: NewConfig(
				":8080", false,
				ModeLocal, "/tmp/users.db", "", "",
				ModeBoth, "/tmp/traffic.db", "https://api.test.com", "",
				60*time.Second, false,
			),
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.validate()

			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}