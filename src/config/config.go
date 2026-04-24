package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Config struct {
	Notifications NotificationsConfig `json:"notifications"`
}

type NotificationsConfig struct {
	Stop       StopConfig       `json:"stop"`
	Permission PermissionConfig `json:"permission"`
}

type StopConfig struct {
	Enabled bool `json:"enabled"`
}

type PermissionConfig struct {
	Enabled bool `json:"enabled"`
}

var defaultConfig = &Config{
	Notifications: NotificationsConfig{
		Stop:       StopConfig{Enabled: true},
		Permission: PermissionConfig{Enabled: true},
	},
}

func getConfigPath() string {
	// Try multiple locations
	paths := []string{
		filepath.Join(os.Getenv("LOCALAPPDATA"), "claude-notifications-win", "config.json"),
		filepath.Join(os.Getenv("APPDATA"), "claude-notifications-win", "config.json"),
		"./config.json",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return paths[0] // Return default path
}

func Load() (*Config, error) {
	path := getConfigPath()

	data, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultConfig, nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
