// Package middleware provides HTTP middleware functions for the Gin router.
package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// Logger logs HTTP requests with structured logging including method, path, status, latency, and client IP.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		c.Next()

		latency := time.Since(startTime)
		statusCode := c.Writer.Status()
		requestID := c.GetString("request_id")

		if raw != "" {
			path = path + "?" + raw
		}

		logEvent := log.Info().
			Str("method", method).
			Str("path", path).
			Int("status", statusCode).
			Dur("latency", latency).
			Str("client_ip", clientIP).
			Str("request_id", requestID)

		if userAgent != "" {
			logEvent = logEvent.Str("user_agent", userAgent)
		}

		if len(c.Errors) > 0 {
			logEvent = logEvent.Str("errors", c.Errors.String())
		}

		if statusCode >= 500 {
			logEvent = log.Error().
				Str("method", method).
				Str("path", path).
				Int("status", statusCode).
				Dur("latency", latency).
				Str("client_ip", clientIP).
				Str("request_id", requestID).
				Str("errors", c.Errors.String())
		} else if statusCode >= 400 {
			logEvent = log.Warn().
				Str("method", method).
				Str("path", path).
				Int("status", statusCode).
				Dur("latency", latency).
				Str("client_ip", clientIP).
				Str("request_id", requestID)
		}

		logEvent.Msg("HTTP request")
	}
}
