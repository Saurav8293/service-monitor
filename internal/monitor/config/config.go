package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ServerAddr          string
	HealthCheckInterval time.Duration
	RequestTimeout      time.Duration
	DatabaseURL         string
	KafkaBroker         string
	KafkaTopic          string
	HealthWorkerCount   int
	JWTSecret           string
	LogLevel            string
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

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func Load() (Config, error) {
	loadEnv(".env")

	serverAddr := getEnv("SERVER_ADDR", ":8080")
	healthCheckInterval := getEnv("HEALTH_CHECK_INTERVAL", "10s")
	requestTimeout := getEnv("HEALTH_REQUEST_TIMEOUT", "5s")
	databaseURL := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/service_monitor?sslmode=disable")
	kafkaBroker := getEnv("KAFKA_BROKER", "localhost:9092")
	kafkaTopic := getEnv("KAFKA_TOPIC", "service-events")
	jwtSecret := getEnv("JWT_SECRET", "super-secret-jwt-key-for-service-monitor")
	logLevel := getEnv("LOG_LEVEL", "info")

	workerCountStr := getEnv("HEALTH_WORKER_COUNT", "5")
	workerCount, err := strconv.Atoi(workerCountStr)
	if err != nil || workerCount <= 0 {
		workerCount = 5
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
		KafkaBroker:         kafkaBroker,
		KafkaTopic:          kafkaTopic,
		HealthWorkerCount:   workerCount,
		JWTSecret:           jwtSecret,
		LogLevel:            logLevel,
	}, nil
}
