package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all application configuration from environment variables.
type Config struct {
	// Database configuration
	DBHost    string
	DBPort    int
	DBUser    string
	DBPass    string
	DBName    string
	DBSSLMode string
	// Server configuration
	ServerPort int
	ServerEnv  string
	// Kafka configuration
	KafkaBrokers     []string
	KafkaTopicEvents string
	// Logging
	LogLevel string
}

// LoadFromEnv loads configuration from environment variables.
func LoadFromEnv() *Config {
	return &Config{
		// Database
		DBHost:    getEnv("DB_HOST", "localhost"),
		DBPort:    getEnvInt("DB_PORT", 5432),
		DBUser:    getEnv("DB_USER", "postgres"),
		DBPass:    getEnv("DB_PASSWORD", "postgres"),
		DBName:    getEnv("DB_NAME", "deus_logistics_db"),
		DBSSLMode: getEnv("DB_SSL_MODE", "disable"),
		// Server
		ServerPort: getEnvInt("SERVER_PORT", 8080),
		ServerEnv:  getEnv("SERVER_ENV", "development"),
		// Kafka
		KafkaBrokers:     []string{getEnv("KAFKA_BROKER", "localhost:9092")},
		KafkaTopicEvents: getEnv("KAFKA_TOPIC_STATUS_CHANGES", "cargo-status-changes"),
		// Logging
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}
}

// DSN returns the PostgreSQL DSN (Data Source Name).
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.DBUser, c.DBPass, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode,
	)
}

// getEnv gets an environment variable with a default fallback.
func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

// getEnvInt gets an integer environment variable with a default fallback.
func getEnvInt(key string, defaultVal int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultVal
}
