package main

import (
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"

	"order-management/gateway"
)

func main() {
	// Load local environment configuration.
	if err := godotenv.Load(); err != nil {
		log.Println(".env not found; using environment variables and defaults")
	}

	cfg, err := gateway.LoadConfig()
	if err != nil {
		log.Fatalf("Gateway configuration error: %v", err)
	}

	handler, err := gateway.NewHandler(cfg)
	if err != nil {
		log.Fatalf("Gateway handler error: %v", err)
	}

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("API Gateway listening on port %s", cfg.Port)
	log.Printf("Order service URL: %s", cfg.OrderServiceURL)

	if err := server.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {
		log.Fatalf("Gateway server failed: %v", err)
	}
}
