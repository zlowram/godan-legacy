package main

import (
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	ServerIP   string
	ServerPort string
	Username   string
	Password   string
	Database   string
}

func loadConfig(configFile string) (Config, error) {
	var config Config
	if _, err := os.Stat(configFile); err != nil {
		return config, err
	}

	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		return config, err
	}

	return config, nil
}
