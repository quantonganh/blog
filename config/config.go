package config

import (
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Navbar struct {
		Items []Item
	}
	SMTP struct {
		Host     string
		Port     int
		Username string
		Password string
	}
	HMAC struct {
		Secret string
	}
}

type Item struct {
	Text string
	URL  string
}

func NewConfig(configPath string) (*Config, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, errors.Errorf("failed to open config file %s: %v", configPath, err)
	}
	defer func() {
		_ = file.Close()
	}()

	config := &Config{}
	if err := yaml.NewDecoder(file).Decode(&config); err != nil {
		return nil, errors.Errorf("failed to decode config: %v", err)
	}

	return config, nil
}
