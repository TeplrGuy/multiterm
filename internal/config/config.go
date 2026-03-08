package config

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/spf13/viper"
)

var validLayouts = []string{"grid", "vertical", "horizontal", "main-side"}

type Config struct {
	Defaults Defaults           `yaml:"defaults" mapstructure:"defaults"`
	Profiles map[string]Profile `yaml:"profiles" mapstructure:"profiles"`
}

type Defaults struct {
	Count  int    `yaml:"count" mapstructure:"count"`
	Layout string `yaml:"layout" mapstructure:"layout"`
}

type Profile struct {
	Count    int      `yaml:"count" mapstructure:"count"`
	Layout   string   `yaml:"layout" mapstructure:"layout"`
	Commands []string `yaml:"commands" mapstructure:"commands"`
}

// DefaultConfig returns a Config populated with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Defaults: Defaults{
			Count:  6,
			Layout: "grid",
		},
		Profiles: make(map[string]Profile),
	}
}

// ConfigPath returns the absolute path to the config file.
func ConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".multiterm.yaml")
	}
	return filepath.Join(home, ".multiterm.yaml")
}

// Load reads configuration from ~/.multiterm.yaml. If the file does not exist
// it returns the default configuration without error.
func Load() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return DefaultConfig(), nil
	}

	viper.SetConfigName(".multiterm")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(home)

	viper.SetDefault("defaults.count", 6)
	viper.SetDefault("defaults.layout", "grid")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if err := validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// GetProfile returns the named profile from cfg. It returns an error if the
// profile does not exist.
func GetProfile(cfg *Config, name string) (*Profile, error) {
	p, ok := cfg.Profiles[name]
	if !ok {
		return nil, fmt.Errorf("profile %q not found", name)
	}
	return &p, nil
}

// InitConfig creates a default ~/.multiterm.yaml if one does not already exist.
func InitConfig() error {
	path := ConfigPath()

	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("config file already exists: %s", path)
	}

	return os.WriteFile(path, []byte(defaultConfigYAML), 0644)
}

func validate(cfg *Config) error {
	if err := validateCount(cfg.Defaults.Count); err != nil {
		return fmt.Errorf("defaults: %w", err)
	}
	if err := validateLayout(cfg.Defaults.Layout); err != nil {
		return fmt.Errorf("defaults: %w", err)
	}

	for name, p := range cfg.Profiles {
		if err := validateCount(p.Count); err != nil {
			return fmt.Errorf("profile %q: %w", name, err)
		}
		if p.Layout != "" {
			if err := validateLayout(p.Layout); err != nil {
				return fmt.Errorf("profile %q: %w", name, err)
			}
		}
	}

	return nil
}

func validateCount(n int) error {
	if n <= 0 || n > 20 {
		return fmt.Errorf("count must be between 1 and 20, got %d", n)
	}
	return nil
}

func validateLayout(layout string) error {
	if !slices.Contains(validLayouts, layout) {
		return fmt.Errorf("invalid layout %q, must be one of %v", layout, validLayouts)
	}
	return nil
}

const defaultConfigYAML = `defaults:
  count: 6
  layout: grid

profiles:
  dev:
    count: 4
    layout: main-side
    commands:
      - ""
      - ""
      - ""
      - ""
  monitor:
    count: 6
    layout: grid
    commands:
      - "htop"
      - ""
      - ""
      - ""
      - ""
      - ""
`
