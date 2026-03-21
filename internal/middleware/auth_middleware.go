package middleware

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shikagari/api/internal/repository/interfaces"
	jwtpkg "github.com/shikagari/api/pkg/jwt"
	"github.com/shikagari/api/pkg/response"
)

const (
	// ContextUserKey is the key under which JWT claims are stored in Gin's context.
	ContextUserKey = "current_user"
)

var (
	errTokenInvalid   = errors.New("token_invalid")
	errSessionMissing = errors.New("session_missing")
	errSessionRevoked = errors.New("session_revoked")
)

// Authenticate is a strict middleware that REQUIRES a valid JWT.
// Use this on routes that must be protected (e.g. create listing, send inquiry).
func Authenticate(jwtManager *jwtpkg.Manager, sessionRepo interfaces.SessionRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			response.Unauthorized(c, "Authentication required. Please log in.")
			c.Abort()
			return
		}

		claims, err := validateClaims(token, jwtManager, sessionRepo)
		if err != nil {
			handleAuthError(c, err)
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
func OptionalAuthenticate(jwtManager *jwtpkg.Manager, sessionRepo interfaces.SessionRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token != "" {
			if claims, err := validateClaims(token, jwtManager, sessionRepo); err == nil {
				c.Set(ContextUserKey, claims)
			}
		}
		c.Next()
	}
}

func validateClaims(token string, jwtManager *jwtpkg.Manager, sessionRepo interfaces.SessionRepository) (*jwtpkg.Claims, error) {
	claims, err := jwtManager.Parse(token)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errTokenInvalid, err)
	}

	if sessionRepo == nil {
		return claims, nil
	}

	if claims.SessionID == uuid.Nil {
		return nil, errSessionMissing
	}

	session, err := sessionRepo.FindByID(claims.SessionID)
	if err != nil {
		return nil, err
	}

	if session == nil || session.UserID != claims.UserID || session.RevokedAt != nil {
		return nil, errSessionRevoked
	}

	_ = sessionRepo.UpdateLastActive(claims.SessionID, time.Now())

	return claims, nil
}

func handleAuthError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, errTokenInvalid):
		response.Unauthorized(c, "Invalid or expired token. Please log in again.")
	case errors.Is(err, errSessionMissing), errors.Is(err, errSessionRevoked):
		response.Unauthorized(c, "Session expired. Please log in again.")
	default:
		response.InternalServerError(c, "Failed to verify session")
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
