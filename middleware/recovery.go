package middleware

import (
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				requestID, _ := c.Get("request_id")

				// Keep panic details in server-side logs only.
				logger.Error("Panic recovered",
					slog.Any("request_id", requestID),
					slog.Any("panic", recovered),
					slog.String("stack_trace", string(debug.Stack())),
				)

				// Stop the remaining handlers for this request.
				c.Abort()

				// Avoid writing a second response if one has already started.
				if !c.Writer.Written() {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": "Internal server error",
					})
				}
			}
		}()

		c.Next()
	}
}
