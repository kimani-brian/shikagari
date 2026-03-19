package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger logs each HTTP request with method, path, status, latency, and client IP.
// Kept lightweight to avoid I/O overhead on high-traffic endpoints.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		if query != "" {
			path = path + "?" + query
		}

		// Colour-code status for readability in development terminals
		statusColour := colourForStatus(statusCode)
		methodColour := colourForMethod(method)
		reset := "\033[0m"

		fmt.Printf("[ShikaGari] %v | %s%3d%s | %12v | %15s | %s%-7s%s %s\n",
			time.Now().Format("2006/01/02 - 15:04:05"),
			statusColour, statusCode, reset,
			latency,
			clientIP,
			methodColour, method, reset,
			path,
		)
	}
}

func colourForStatus(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "\033[32m" // green
	case code >= 300 && code < 400:
		return "\033[36m" // cyan
	case code >= 400 && code < 500:
		return "\033[33m" // yellow
	default:
		return "\033[31m" // red
	}
}

func colourForMethod(method string) string {
	switch method {
	case "GET":
		return "\033[34m" // blue
	case "POST":
		return "\033[32m" // green
	case "PUT", "PATCH":
		return "\033[33m" // yellow
	case "DELETE":
		return "\033[31m" // red
	default:
		return "\033[37m" // white
	}
}
