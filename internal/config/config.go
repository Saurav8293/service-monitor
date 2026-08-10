package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	ServerAddr          string
	HealthCheckInterval time.Duration
	RequestTimeout      time.Duration
	DatabaseURL         string
}

func loadEnv(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			if os.Getenv(key) == "" {
				os.Setenv(key, val)
			}
		}
	}
}

func Load() (Config, error) {
	loadEnv(".env")

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
