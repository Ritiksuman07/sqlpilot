package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Profile struct {
	Name     string `yaml:"name"`
	Driver   string `yaml:"driver"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"-"`
	Path     string `yaml:"path"`
}

type Config struct {
	Profiles []Profile `yaml:"profiles"`
}

func (c Config) FindProfile(name string) (Profile, bool) {
	for _, profile := range c.Profiles {
		if profile.Name == name {
			return profile, true
		}
	}
	return Profile{}, false
}

func BuildDSN(profile Profile, password string) string {
	switch profile.Driver {
	case "postgres":
		if profile.Port == 0 {
			profile.Port = 5432
		}
		return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", profile.Username, password, profile.Host, profile.Port, profile.Database)
	case "mysql":
		if profile.Port == 0 {
			profile.Port = 3306
		}
		return fmt.Sprintf("mysql://%s:%s@%s:%d/%s", profile.Username, password, profile.Host, profile.Port, profile.Database)
	case "sqlite":
		return profile.Path
	case "duckdb":
		return profile.Path
	default:
		return ""
	}
}

func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "sqlpilot", "connections.yaml"), nil
}

func Load() (Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return Config{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, nil
		}
		return Config{}, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func Save(cfg Config) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
