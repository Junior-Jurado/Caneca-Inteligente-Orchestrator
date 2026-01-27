// Package middleware provides HTTP middleware functions for the Gin router.
package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders adds security-related HTTP headers to prevent common web vulnerabilities
// including clickjacking, XSS attacks, MIME sniffing, and information disclosure.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")

		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		c.Header("Content-Security-Policy", "default-src 'none'")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("Permissions-Policy",
			"geolocation=(), microphone=(), camera=(), payment=()")
		c.Header("X-Powered-By", "")

		c.Next()
	}
}
