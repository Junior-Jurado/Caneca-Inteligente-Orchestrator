// Package middleware provides HTTP middleware functions for the Gin router.
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID generates or propagates a unique request ID for each HTTP request.
// The request ID is available to handlers via the context and is included in response headers.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")

		if requestID == "" {
			requestID = generateRequestID()
		}

		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// generateRequestID creates a new unique request ID with format "req_" followed by 8 random characters.
func generateRequestID() string {
	id := uuid.New().String()
	return "req_" + id[:8]
}
