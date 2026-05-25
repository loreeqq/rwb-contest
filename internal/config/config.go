package config

import (
	"os"
	"strconv"
)

const WindowSeconds = 300

type Config struct {
	HTTPAddr     string
	RabbitURL    string
	RabbitQueue  string
	WorkersCount int
}

func Load() Config {
	return Config{
		HTTPAddr:     getEnv("HTTP_ADDR", ":8080"),
		RabbitURL:    getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		RabbitQueue:  getEnv("RABBITMQ_QUEUE", "search_queries"),
		WorkersCount: getEnvAsInt("WORKERS_COUNT", 4),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	return value
}

func getEnvAsInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return defaultValue
	}

	return parsed
}
