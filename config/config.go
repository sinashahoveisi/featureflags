package config

import (
	"os"
	"strconv"
	"time"
)

type Application struct {
	GracefulShutdownTimeout time.Duration
}

type HTTPServer struct {
	Port int
}

type Database struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

type Logger struct {
	Level string
	Mode  string // development or production
}

type Swagger struct {
	Enabled bool `json:"enabled"`
}

type Config struct {
	Application Application
	HTTPServer  HTTPServer
	Database    Database
	Logger      Logger
	Swagger     Swagger
}

func Load() (*Config, error) {
	cfg := &Config{
		Application: Application{
			GracefulShutdownTimeout: parseDurationWithDefault("APPLICATION_GRACEFUL_SHUTDOWN_TIMEOUT", 30*time.Second),
		},
		HTTPServer: HTTPServer{
			Port: parseIntWithDefault("HTTP_SERVER_PORT", 8080),
		},
		Database: Database{
			Host:     getEnvWithDefault("DATABASE_HOST", "db"),
			Port:     parseIntWithDefault("DATABASE_PORT", 5432),
			User:     getEnvWithDefault("DATABASE_USER", "featureflags"),
			Password: getEnvWithDefault("DATABASE_PASSWORD", "featureflags"),
			Name:     getEnvWithDefault("DATABASE_NAME", "featureflags"),
			SSLMode:  getEnvWithDefault("DATABASE_SSL_MODE", "disable"),
		},
		Logger: Logger{
			Level: getEnvWithDefault("LOGGER_LEVEL", "info"),
			Mode:  getEnvWithDefault("LOGGER_MODE", "production"),
		},
	}

	// Set Swagger defaults
	cfg.Swagger = Swagger{
		Enabled: getEnvBoolWithDefault("SWAGGER_ENABLED", true),
	}

	// Support legacy environment variables
	if port := os.Getenv("APP_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.HTTPServer.Port = p
		}
	}
	if host := os.Getenv("POSTGRES_HOST"); host != "" {
		cfg.Database.Host = host
	}
	if user := os.Getenv("POSTGRES_USER"); user != "" {
		cfg.Database.User = user
	}
	if password := os.Getenv("POSTGRES_PASSWORD"); password != "" {
		cfg.Database.Password = password
	}
	if name := os.Getenv("POSTGRES_DB"); name != "" {
		cfg.Database.Name = name
	}

	return cfg, nil
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseIntWithDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func parseDurationWithDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvBoolWithDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
} 