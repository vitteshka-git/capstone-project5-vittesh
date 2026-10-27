package gateway

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Port            string
	OrderServiceURL string
}

func LoadConfig() (Config, error) {
	port := strings.TrimSpace(os.Getenv("GATEWAY_PORT"))
	if port == "" {
		port = "8080"
	}

	orderURL := strings.TrimRight(
		strings.TrimSpace(os.Getenv("ORDER_SERVICE_URL")),
		"/",
	)
	if orderURL == "" {
		orderURL = "http://localhost:8081"
	}

	if !strings.HasPrefix(orderURL, "http://") &&
		!strings.HasPrefix(orderURL, "https://") {
		return Config{}, fmt.Errorf("ORDER_SERVICE_URL must be an HTTP or HTTPS URL")
	}

	return Config{
		Port:            port,
		OrderServiceURL: orderURL,
	}, nil
}