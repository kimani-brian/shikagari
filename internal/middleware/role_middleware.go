package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/shikagari/api/internal/domain"
	jwtpkg "github.com/shikagari/api/pkg/jwt"
	"github.com/shikagari/api/pkg/response"
)

// RequireRole returns a middleware that allows access only to users
// whose role matches one of the provided allowed roles.
//
// Usage:
//
//	router.PATCH("/listings/:id", middleware.RequireRole(domain.RoleDealer, domain.RoleAdmin), handler)
func RequireRole(allowedRoles ...domain.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetCurrentUser(c)
		if claims == nil {
			response.Unauthorized(c, "Authentication required.")
			c.Abort()
			return
		}

		for _, role := range allowedRoles {
			if claims.Role == role {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "You do not have permission to access this resource.")
		c.Abort()
	}
}

// RequireAdmin is a convenience wrapper for admin-only routes.
func RequireAdmin() gin.HandlerFunc {
	return RequireRole(domain.RoleAdmin)
}

// RequireDealer allows access to business dealership accounts and admins.
func RequireDealer() gin.HandlerFunc {
	return RequireRole(domain.RoleDealer, domain.RoleAdmin)
}

// RequireVerified ensures only users with the verified badge can proceed.
// This is used for actions exclusive to verified sellers.
func RequireVerified() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetCurrentUser(c)
		if claims == nil {
			response.Unauthorized(c, "Authentication required.")
			c.Abort()
			return
		}

		if !claims.IsVerified {
			response.Forbidden(c, "A verified badge is required for this action.")
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetCurrentUser retrieves JWT claims from the Gin context.
// Returns nil if the user is not authenticated (i.e. on public routes).
func GetCurrentUser(c *gin.Context) *jwtpkg.Claims {
	val, exists := c.Get(ContextUserKey)
	if !exists {
		return nil
	}

	claims, ok := val.(*jwtpkg.Claims)
	if !ok {
		return nil
	}

	return claims
}
