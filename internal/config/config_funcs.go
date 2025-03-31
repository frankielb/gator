package config

import (
	"encoding/json"
	"os"
)

const configPath = "/.gatorconfig.json"

type Config struct {
	DbURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func Read() (Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return Config{}, err
	}
	jsonpath := homeDir + configPath

	file, err := os.Open(jsonpath)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return Config{}, err
	}
	return config, nil
}

func write(cfg Config) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	jsonpath := homeDir + configPath

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(jsonpath, data, 0644)

}

func (c *Config) SetUser(userName string) error {
	c.CurrentUserName = userName
	return write(*c)
}
