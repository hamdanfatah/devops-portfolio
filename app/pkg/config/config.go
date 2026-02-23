package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

// Config holds all application configuration
type Config struct {
	// Server
	ServerPort string `envconfig:"SERVER_PORT" default:"8080"`
	Environment string `envconfig:"ENVIRONMENT" default:"development"`

	// PostgreSQL
	PostgresHost     string `envconfig:"POSTGRES_HOST" default:"localhost"`
	PostgresPort     string `envconfig:"POSTGRES_PORT" default:"5432"`
	PostgresUser     string `envconfig:"POSTGRES_USER" default:"hamfa"`
	PostgresPassword string `envconfig:"POSTGRES_PASSWORD" default:"hamfa_secret"`
	PostgresDB       string `envconfig:"POSTGRES_DB" default:"taskmanager"`
	PostgresSSLMode  string `envconfig:"POSTGRES_SSL_MODE" default:"disable"`

	// MongoDB
	MongoURI string `envconfig:"MONGO_URI" default:"mongodb://localhost:27017"`
	MongoDB  string `envconfig:"MONGO_DB" default:"taskmanager"`

	// Redis
	RedisHost     string `envconfig:"REDIS_HOST" default:"localhost"`
	RedisPort     string `envconfig:"REDIS_PORT" default:"6379"`
	RedisPassword string `envconfig:"REDIS_PASSWORD" default:""`
	RedisDB       int    `envconfig:"REDIS_DB" default:"0"`
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return &cfg, nil
}

// PostgresDSN returns the PostgreSQL connection string
func (c *Config) PostgresDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.PostgresUser, c.PostgresPassword, c.PostgresHost,
		c.PostgresPort, c.PostgresDB, c.PostgresSSLMode,
	)
}

// RedisAddr returns the Redis address
func (c *Config) RedisAddr() string {
	return fmt.Sprintf("%s:%s", c.RedisHost, c.RedisPort)
}
