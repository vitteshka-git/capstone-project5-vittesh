package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const RequestIDHeader = "X-Request-ID"

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := strings.TrimSpace(c.GetHeader(RequestIDHeader))

		// Generate an ID if the client did not provide one.
		if requestID == "" {
			requestID = uuid.NewString()
		}

		// Make the ID available to the client and application.
		c.Header(RequestIDHeader, requestID)
		c.Set("request_id", requestID)

		c.Next()
	}
}
