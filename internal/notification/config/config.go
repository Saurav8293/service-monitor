package config

import (
	"bufio"
	"os"
	"strings"
)

type Config struct {
	GRPCPort    string
	MetricsAddr string
	DatabaseURL string
	LogLevel    string
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

func Load() Config {
	loadEnv(".env")

	return Config{
		GRPCPort:    getEnv("NOTIFICATION_GRPC_PORT", ":50051"),
		MetricsAddr: getEnv("NOTIFICATION_METRICS_ADDR", ":8084"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/service_monitor?sslmode=disable"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}
}
