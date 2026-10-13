package config

import (
	"bufio"
	"os"
	"strings"
)

type Config struct {
	MetricsAddr     string
	DatabaseURL     string
	KafkaBroker     string
	KafkaTopic      string
	KafkaGroupID    string
	NotificationURL string // gRPC address: e.g. "notification-service:50051" or "localhost:50051"
	LogLevel        string
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
		MetricsAddr:     getEnv("METRICS_ADDR", ":8083"),
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/service_monitor?sslmode=disable"),
		KafkaBroker:     getEnv("KAFKA_BROKER", "localhost:9092"),
		KafkaTopic:      getEnv("KAFKA_TOPIC", "service-events"),
		KafkaGroupID:    getEnv("KAFKA_GROUP_ID", "incident-service-group"),
		NotificationURL: getEnv("NOTIFICATION_GRPC_ADDR", "localhost:50051"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
	}
}
