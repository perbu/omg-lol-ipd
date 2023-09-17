package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	ApiKey   string `json:"api_key"`
	Hostname string `json:"hostname"`
	Domain   string `json:"domain"`
}

func Load(file string) (Config, error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return Config{}, fmt.Errorf("os.ReadFile(%s): %w", file, err)
	}
	var config Config
	err = json.Unmarshal(content, &config)
	if err != nil {
		return Config{}, fmt.Errorf("json.Unmarshal: %w", err)
	}
	return config, nil
}
