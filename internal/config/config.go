package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Database DatabaseConfig
	RabbitMQ RabbitMQConfig
	App      AppConfig
}

type DatabaseConfig struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
}

type RabbitMQConfig struct {
	ConnectionUrl  string
	IngestionQueue string
}

type AppConfig struct {
	BatchIngestionSize int
	BatchFlushInterval int
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found, using environment variables")
	}

	config := &Config{
		Database: DatabaseConfig{
			Host:     getEnv("CLICKHOUSE_HOST", "localhost"),
			Port:     getEnvAsInt("CLICKHOUSE_PORT", 9000),
			Database: getEnv("CLICKHOUSE_DATABASE", "audit-ingestion-service"),
			Username: getEnv("CLICKHOUSE_USERNAME", "default"),
			Password: getEnv("CLICKHOUSE_PASSWORD", ""),
		},
		RabbitMQ: RabbitMQConfig{
			ConnectionUrl:  getEnv("RABBITMQ_CONNECTION_URL", "amqp://localhost"),
			IngestionQueue: getEnv("RABBITMQ_INGESTION_QUEUE_NAME", "audit-ingestion"),
		},
		App: AppConfig{
			BatchIngestionSize: getEnvAsInt("BATCH_INGESTION_SIZE", 100),
			BatchFlushInterval: getEnvAsInt("BATCH_FLUSH_INTERVAL_IN_SECONDS", 30),
		},
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if valueStr := os.Getenv(key); valueStr != "" {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}

	return defaultValue
}
