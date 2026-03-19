package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	jwtpkg "github.com/shikagari/api/pkg/jwt"
	"github.com/shikagari/api/pkg/response"
)

const (
	// ContextUserKey is the key under which JWT claims are stored in Gin's context.
	ContextUserKey = "current_user"
)

// Authenticate is a strict middleware that REQUIRES a valid JWT.
// Use this on routes that must be protected (e.g. create listing, send inquiry).
func Authenticate(jwtManager *jwtpkg.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			response.Unauthorized(c, "Authentication required. Please log in.")
			c.Abort()
			return
		}

		claims, err := jwtManager.Parse(token)
		if err != nil {
			response.Unauthorized(c, "Invalid or expired token. Please log in again.")
			c.Abort()
			return
		}

		// Store claims in context for downstream handlers
		c.Set(ContextUserKey, claims)
		c.Next()
	}
}

// OptionalAuthenticate is a soft middleware that parses the JWT if present
// but does NOT block the request if the token is absent or invalid.
// Use this on public routes where auth enriches (but is not required for) the response.
func OptionalAuthenticate(jwtManager *jwtpkg.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token != "" {
			if claims, err := jwtManager.Parse(token); err == nil {
				c.Set(ContextUserKey, claims)
			}
		}
		c.Next()
	}
}

// extractToken pulls the Bearer token from the Authorization header.
// Expected header format: "Authorization: Bearer <token>"
func extractToken(c *gin.Context) string {
	header := c.GetHeader("Authorization")
	if header == "" {
		return ""
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return ""
	}

	return strings.TrimSpace(parts[1])
}
