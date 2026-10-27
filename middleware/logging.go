package middleware

import (
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func Logging() gin.HandlerFunc {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	return func(c *gin.Context) {
		start := time.Now()

		// Process the request through the remaining middleware and handler.
		c.Next()

		duration := time.Since(start)

		requestID, _ := c.Get("request_id")

		// Record structured fields as a JSON log entry.
		logger.Info("HTTP request completed",
			slog.Any("request_id", requestID),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("duration", duration),
			slog.String("client_ip", c.ClientIP()),
		)
	}
}
