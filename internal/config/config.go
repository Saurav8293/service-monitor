package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	ServerAddr          string
	HealthCheckInterval time.Duration
	RequestTimeout      time.Duration
	DatabaseURL         string
}

func Load() (Config, error) {
	serverAddr := os.Getenv("SERVER_ADDR")
	healthCheckInterval := os.Getenv("HEALTH_CHECK_INTERVAL")
	requestTimeout := os.Getenv("HEALTH_REQUEST_TIMEOUT")
	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	if serverAddr == "" {
		return Config{}, fmt.Errorf("SERVER_ADDR is required")
	}

	interval, err := time.ParseDuration(healthCheckInterval)
	if err != nil {
		return Config{}, fmt.Errorf("invalid HEALTH_CHECK_INTERVAL: %w", err)
	}

	timeout, err := time.ParseDuration(requestTimeout)
	if err != nil {
		return Config{}, fmt.Errorf("invalid HEALTH_REQUEST_TIMEOUT: %w", err)
	}

	return Config{
		ServerAddr:          serverAddr,
		HealthCheckInterval: interval,
		RequestTimeout:      timeout,
		DatabaseURL:         databaseURL,
	}, nil

}
