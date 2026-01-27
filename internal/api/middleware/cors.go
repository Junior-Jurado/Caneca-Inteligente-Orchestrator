// Package middleware provides HTTP middleware functions for the Gin router.
// It includes CORS, logging, authentication, request ID tracking, and security headers.
package middleware

import (
	"github.com/gin-gonic/gin"
)

// CORS enables Cross-Origin Resource Sharing for API endpoints.
// It allows browsers to make requests from different origins (domains).
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers",
			"Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, "+
				"Authorization, accept, origin, Cache-Control, X-Requested-With, "+
				"X-API-KEY, X-Request-ID")
		c.Header("Access-Control-Allow-Methods",
			"POST, GET, OPTIONS, PUT, DELETE, PATCH")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
