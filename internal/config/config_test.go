package config

import (
	"os"
	"testing"
)

func TestLoad_NoConfigFile(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() with no config file should not error: %v", err)
	}

	if cfg.Defaults.Count != 6 {
		t.Errorf("default count = %d, want 6", cfg.Defaults.Count)
	}
	if cfg.Defaults.Layout != "grid" {
		t.Errorf("default layout = %q, want 'grid'", cfg.Defaults.Layout)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Defaults.Count != 6 {
		t.Errorf("default count = %d, want 6", cfg.Defaults.Count)
	}
	if cfg.Defaults.Layout != "grid" {
		t.Errorf("default layout = %q, want 'grid'", cfg.Defaults.Layout)
	}
	if cfg.Profiles == nil {
		t.Error("profiles map should not be nil")
	}
}

func TestConfigPath(t *testing.T) {
	path := ConfigPath()
	if path == "" {
		t.Error("ConfigPath returned empty string")
	}
}

func TestGetProfile_Existing(t *testing.T) {
	cfg := &Config{
		Profiles: map[string]Profile{
			"dev": {Count: 4, Layout: "main-side", Commands: []string{"npm start"}},
		},
	}

	profile, err := GetProfile(cfg, "dev")
	if err != nil {
		t.Fatalf("GetProfile should find 'dev': %v", err)
	}
	if profile.Count != 4 {
		t.Errorf("profile count = %d, want 4", profile.Count)
	}
	if profile.Layout != "main-side" {
		t.Errorf("profile layout = %q, want 'main-side'", profile.Layout)
	}
}

func TestGetProfile_Missing(t *testing.T) {
	cfg := &Config{
		Profiles: map[string]Profile{},
	}

	_, err := GetProfile(cfg, "nonexistent")
	if err == nil {
		t.Error("GetProfile should error for nonexistent profile")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{"valid config", &Config{Defaults: Defaults{Count: 4, Layout: "grid"}}, false},
		{"count too low", &Config{Defaults: Defaults{Count: 0, Layout: "grid"}}, true},
		{"count too high", &Config{Defaults: Defaults{Count: 21, Layout: "grid"}}, true},
		{"invalid layout", &Config{Defaults: Defaults{Count: 4, Layout: "invalid"}}, true},
		{"vertical", &Config{Defaults: Defaults{Count: 4, Layout: "vertical"}}, false},
		{"horizontal", &Config{Defaults: Defaults{Count: 4, Layout: "horizontal"}}, false},
		{"main-side", &Config{Defaults: Defaults{Count: 4, Layout: "main-side"}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate(tt.cfg)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateCount(t *testing.T) {
	for _, n := range []int{1, 5, 10, 20} {
		if err := validateCount(n); err != nil {
			t.Errorf("validateCount(%d) should pass: %v", n, err)
		}
	}
	for _, n := range []int{0, -1, 21, 100} {
		if err := validateCount(n); err == nil {
			t.Errorf("validateCount(%d) should fail", n)
		}
	}
}

func TestValidateLayout(t *testing.T) {
	for _, l := range []string{"grid", "vertical", "horizontal", "main-side"} {
		if err := validateLayout(l); err != nil {
			t.Errorf("validateLayout(%q) should pass: %v", l, err)
		}
	}
	for _, l := range []string{"invalid", "", "tile", "auto"} {
		if err := validateLayout(l); err == nil {
			t.Errorf("validateLayout(%q) should fail", l)
		}
	}
}

func TestMarshalConfig(t *testing.T) {
	cfg := &Config{
		Defaults: Defaults{Count: 4, Layout: "grid"},
		Profiles: map[string]Profile{
			"test": {Count: 3, Layout: "vertical", Commands: []string{"echo hi"}},
		},
	}

	data, err := marshalConfig(cfg)
	if err != nil {
		t.Fatalf("marshalConfig failed: %v", err)
	}

	content := string(data)
	if content == "" {
		t.Error("marshaled config is empty")
	}
}

func TestSaveProfile(t *testing.T) {
	// Save to home dir config, so use a temp HOME.
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg := DefaultConfig()
	profile := Profile{
		Count:    3,
		Layout:   "vertical",
		Commands: []string{"echo hello"},
	}

	if err := SaveProfile(cfg, "test-profile", profile); err != nil {
		t.Fatalf("SaveProfile failed: %v", err)
	}

	// Verify the profile is in cfg.
	if _, ok := cfg.Profiles["test-profile"]; !ok {
		t.Error("profile not added to config")
	}
}
