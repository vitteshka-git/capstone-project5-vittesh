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
	AppPort                  string
	LogLevel                 string
	DBHost                   string
	DBPort                   string
	DBName                   string
	DBUser                   string
	DBPassword               string
	DBSSLMode                string
	ProductServiceURL        string
	ProductServiceTimeout    time.Duration
	ProductServiceMaxRetries int
	AWSRegion                string
	AWSSecretName            string
}

func Load() (*Config, error) {
	// Load .env for local development.
	// In AWS, environment variables can be injected by the deployment.
	_ = godotenv.Load()

	cfg := &Config{
		AppPort:           getEnv("APP_PORT", "8080"),
		LogLevel:          strings.ToLower(getEnv("LOG_LEVEL", "info")),
		DBHost:            getEnv("DB_HOST", ""),
		DBPort:            getEnv("DB_PORT", "5432"),
		DBName:            getEnv("DB_NAME", "orderdb"),
		DBUser:            getEnv("DB_USER", ""),
		DBPassword:        os.Getenv("DB_PASSWORD"),
		DBSSLMode:         getEnv("DB_SSLMODE", "require"),
		ProductServiceURL: getEnv("PRODUCT_SERVICE_URL", "http://localhost:8081"),
		AWSRegion:         getEnv("AWS_REGION", "eu-north-1"),
		AWSSecretName:     getEnv("AWS_SECRET_NAME", ""),
	}

	timeout, err := time.ParseDuration(
		getEnv("PRODUCT_SERVICE_TIMEOUT", "2s"),
	)
	if err != nil || timeout <= 0 {
		return nil, fmt.Errorf("PRODUCT_SERVICE_TIMEOUT must be a positive duration")
	}
	cfg.ProductServiceTimeout = timeout

	retries, err := strconv.Atoi(
		getEnv("PRODUCT_SERVICE_MAX_RETRIES", "3"),
	)
	if err != nil || retries < 0 || retries > 5 {
		return nil, fmt.Errorf("PRODUCT_SERVICE_MAX_RETRIES must be between 0 and 5")
	}
	cfg.ProductServiceMaxRetries = retries

	if cfg.AppPort == "" {
		return nil, fmt.Errorf("APP_PORT cannot be empty")
	}

	if cfg.DBHost == "" || cfg.DBPort == "" ||
		cfg.DBName == "" || cfg.DBUser == "" {
		return nil, fmt.Errorf(
			"DB_HOST, DB_PORT, DB_NAME, and DB_USER are required",
		)
	}

	if cfg.DBPassword == "" && cfg.AWSSecretName == "" {
		return nil, fmt.Errorf(
			"set DB_PASSWORD or configure AWS_SECRET_NAME for secret retrieval",
		)
	}

	if cfg.DBSSLMode != "disable" &&
		cfg.DBSSLMode != "require" &&
		cfg.DBSSLMode != "verify-ca" &&
		cfg.DBSSLMode != "verify-full" {
		return nil, fmt.Errorf("invalid DB_SSLMODE")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
