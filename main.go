
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"order-management/config"
	"order-management/controller"
	"order-management/database"
	"order-management/graphqlapi"
	"order-management/middleware"
	"order-management/model"
	"order-management/product"
	"order-management/repository"
	"order-management/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to PostgreSQL using GORM.
	ctx, cancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)

	db, err := database.Connect(ctx, cfg)
	cancel()

	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database connection: %v", err)
	}
	defer sqlDB.Close()

	log.Println("PostgreSQL connected successfully")

	// Create or update the orders table automatically.
	migrationCtx, migrationCancel := context.WithTimeout(
		context.Background(),
		2*time.Minute,
	)

	err = db.WithContext(migrationCtx).AutoMigrate(&model.Order{})
	migrationCancel()

	if err != nil {
		log.Fatalf("Failed to migrate database schema: %v", err)
	}

	log.Println("Database migration completed successfully")

	// Initialize the Product Service client.
	productClient, err := product.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create product client: %v", err)
	}

	// Initialize repository, service, and controller layers.
	orderRepo := repository.NewOrderRepository(db)
	orderService := service.NewOrderService(orderRepo, productClient)
	orderController := controller.NewOrderController(orderService)

	// Initialize the GraphQL schema.
	graphqlSchema, err := graphqlapi.NewSchema(orderService)
	if err != nil {
		log.Fatalf("Failed to create GraphQL schema: %v", err)
	}

	// Create the Gin router.
	router := gin.New()

	router.Use(
		middleware.RequestID(),
		middleware.Logging(),
		middleware.Recovery(),
	)

	// Liveness endpoint.
	router.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "alive",
		})
	})

	// Readiness endpoint checks database connectivity.
	router.GET("/health/ready", func(c *gin.Context) {
		healthCtx, healthCancel := context.WithTimeout(
			c.Request.Context(),
			3*time.Second,
		)
		defer healthCancel()

		if err := database.CheckHealth(healthCtx, db); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not ready",
				"error":  "database unavailable",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
		})
	})

	// Public login endpoint.
	router.POST("/api/v1/auth/login", middleware.Login)

	// GraphQL endpoint. A valid JWT is required.
	router.POST(
		"/graphql",
		middleware.RequireAuth(),
		graphqlapi.Handler(graphqlSchema),
	)

	// Order REST API routes.
	api := router.Group("/api/v1")
	{
		api.POST(
			"/orders",
			middleware.RequireAuth(),
			orderController.CreateOrder,
		)

		api.GET("/orders", orderController.ListOrders)
		api.GET("/orders/:id", orderController.GetOrder)

		api.PUT(
			"/orders/:id/status",
			middleware.RequireAuth(),
			orderController.UpdateOrderStatus,
		)
	}

	// Configure the HTTP server.
	port := cfg.AppPort
	if port == "" {
		port = "8081"
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Start the Order Management API.
	go func() {
		log.Printf("Order Management API listening on port %s", port)

		if err := server.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server failed: %v", err)
		}
	}()

	// Wait for an interrupt or termination signal.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down Order Management API...")

	// Gracefully shut down the server.
	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Graceful shutdown failed: %v", err)
	}
}
