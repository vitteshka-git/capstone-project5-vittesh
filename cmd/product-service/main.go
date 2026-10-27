
package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Product struct {
	ProductID int64   `json:"productId"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
}

var products = map[int64]Product{
	1: {ProductID: 1, Name: "Keyboard", Price: 50.00},
	2: {ProductID: 2, Name: "Mouse", Price: 25.00},
	3: {ProductID: 3, Name: "Monitor", Price: 200.00},
}

func main() {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "alive"})
	})

	router.GET("/api/v1/products/:id", func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil || id <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid product ID",
			})
			return
		}

		product, ok := products[id]
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Product not found",
			})
			return
		}

		c.JSON(http.StatusOK, product)
	})

	port := strings.TrimSpace("8082")
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("Product Service listening on port %s", port)
	if err := server.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {
		log.Fatalf("Product Service failed: %v", err)
	}
}
