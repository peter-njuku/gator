package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DbURL           string `json:"db_url"`
	CurrentUsername string `json:"username"`
}

func (cfg *Config) SetUser(Username string) error {
	cfg.CurrentUsername = Username
	return write(*cfg)
}

func write(cfg Config) error {
	path, err := getConfigFile()
	if err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("unable to create config file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err = encoder.Encode(cfg); err != nil {
		return fmt.Errorf("unable to encode config to file: %w", err)
	}

	return nil
}

func getConfigFile() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to locate home directory: %w", err)
	}

	return filepath.Join(homeDir, configFileName), nil
}

func Read() (Config, error) {
	var cfg Config

	configPath, err := getConfigFile()
	if err != nil {
		return cfg, err
	}

	file, err := os.Open(configPath)
	if err != nil {
		return cfg, fmt.Errorf("unable to open config file: %w", err)
	}

	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	if err != nil {
		return cfg, fmt.Errorf("unable to decode config file: %w", err)
	}

	return cfg, nil
}
