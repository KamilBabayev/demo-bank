package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	DefaultAPIURL = "http://localhost:8080/api/v1"
	configDir     = ".dbank"
	configFile    = "config.json"
)

type Config struct {
	Token    string `json:"token,omitempty"`
	APIURL   string `json:"api_url,omitempty"`
	Username string `json:"username,omitempty"`
	UserID   int64  `json:"user_id,omitempty"`
	Role     string `json:"role,omitempty"`
}

func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, configDir, configFile), nil
}

func Load() (*Config, error) {
	path, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		APIURL: DefaultAPIURL,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	if cfg.APIURL == "" {
		cfg.APIURL = DefaultAPIURL
	}

	return cfg, nil
}

func (c *Config) Save() error {
	path, err := getConfigPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func (c *Config) Clear() error {
	c.Token = ""
	c.Username = ""
	c.UserID = 0
	c.Role = ""
	return c.Save()
}

func (c *Config) IsLoggedIn() bool {
	return c.Token != ""
}
