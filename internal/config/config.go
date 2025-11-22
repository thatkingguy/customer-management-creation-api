package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	API      APIConfig
	Logger   LoggerConfig
}

type ServerConfig struct {
	Port string
	Env  string
}

type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type APIConfig struct {
	Key string
}

type LoggerConfig struct {
	Level  string
	Format string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (optional, won't fail if missing)
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Env:  getEnvWithFallback("ENV", "NODE_ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", ""),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", ""),
			Password:        getEnvWithFallback("DB_PASSWORD", "DB_PWD", ""),
			Name:            getEnv("DB_NAME", ""),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    getEnvAsIntWithFallback("DB_MAX_OPEN_CONNS", "DB_MAX_POOL_SIZE", 25),
			MaxIdleConns:    getEnvAsIntWithFallback("DB_MAX_IDLE_CONNS", "DB_MIN_POOL_SIZE", 5),
			ConnMaxLifetime: time.Duration(getEnvAsInt("DB_CONN_MAX_LIFETIME", 5)) * time.Minute,
		},
		API: APIConfig{
			Key: getEnv("SMARTADAPTER_API_KEY", ""),
		},
		Logger: LoggerConfig{
			Level:  getLogLevel(),
			Format: getEnv("LOG_FORMAT", "text"),
		},
	}

	// Validate required fields
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Validate checks that all required configuration values are set
func (c *Config) Validate() error {
	if c.Database.Host == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("DB_USER is required")
	}
	if c.Database.Password == "" {
		return fmt.Errorf("DB_PASSWORD is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("DB_NAME is required")
	}
	if c.API.Key == "" {
		return fmt.Errorf("SMARTADAPTER_API_KEY is required")
	}
	return nil
}

// ConnectionString returns the PostgreSQL connection string
func (c *DatabaseConfig) ConnectionString() string {
	// Add connection timeout and performance optimizations
	// connect_timeout: Fail fast if can't connect (default 0 = wait forever)
	// synchronous_commit=off: Disable synchronous commit for better performance (data is still safe, just not immediately flushed)
	// Note: This trades off some durability guarantees for performance. For production, consider using 'local' or 'remote_write'
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s connect_timeout=5 synchronous_commit=off",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvWithFallback tries the primary key first, then falls back to the fallback key
func getEnvWithFallback(primaryKey, fallbackKey, defaultValue string) string {
	if value := os.Getenv(primaryKey); value != "" {
		return value
	}
	if value := os.Getenv(fallbackKey); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

// getEnvAsIntWithFallback tries the primary key first, then falls back to the fallback key
func getEnvAsIntWithFallback(primaryKey, fallbackKey string, defaultValue int) int {
	if valueStr := os.Getenv(primaryKey); valueStr != "" {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}
	if valueStr := os.Getenv(fallbackKey); valueStr != "" {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}

// getLogLevel determines log level based on environment
// Priority: LOG_LEVEL env var > environment-based default
// Dev/Development: debug (verbose)
// QA/Staging: info
// Prod/Production: info (only important logs)
func getLogLevel() string {
	// If LOG_LEVEL is explicitly set, use it (allows override)
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		return level
	}

	// Set default based on ENV/NODE_ENV
	env := getEnvWithFallback("ENV", "NODE_ENV", "development")
	envLower := strings.ToLower(env)
	switch envLower {
	case "development", "dev":
		return "debug"
	case "staging", "qa", "test":
		return "info"
	case "production", "prod":
		return "info"
	default:
		// Default to info for safety (not too verbose)
		return "info"
	}
}
