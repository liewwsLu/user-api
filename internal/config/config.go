package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DatabaseURL string
	ServerPort  int
}

func Load() (Config, error) {
	const defaultServerPort = 8080
	cfg := Config{
		ServerPort: defaultServerPort,
	}
	databaseURL, databaseExists := os.LookupEnv("DATABASE_URL")
	databaseURL = strings.TrimSpace(databaseURL)
	if !databaseExists || databaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	cfg.DatabaseURL = databaseURL
	serverPortText, serverPortExists := os.LookupEnv("SERVER_PORT")
	if !serverPortExists {
		return cfg, nil
	}
	serverPortText = strings.TrimSpace(serverPortText)
	if serverPortText == "" {
		return cfg, nil
	}
	serverPort, err := strconv.Atoi(serverPortText)
	if err != nil {
		return Config{}, errors.New("SERVER_PORT must be an integer")
	}
	if serverPort < 1 || serverPort > 65535 {
		return Config{}, errors.New("SERVER_PORT must be between 1 and 65535")
	}
	cfg.ServerPort = serverPort
	return cfg, nil
}
