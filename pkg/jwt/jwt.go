package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/shikagari/api/internal/domain"
)

// Claims defines the payload embedded in every ShikaGari JWT.
// We include only the minimum fields needed for auth decisions
// to keep token size small (important for mobile/low-bandwidth users).
type Claims struct {
	UserID     uuid.UUID       `json:"user_id"`
	Email      string          `json:"email"`
	Role       domain.UserRole `json:"role"`
	IsVerified bool            `json:"is_verified"`
	jwt.RegisteredClaims
}

// Manager handles JWT generation and validation.
type Manager struct {
	secret      []byte
	expiryHours time.Duration
}

// New returns a JWT Manager configured with the application secret
// and token expiry duration.
func New(secret string, expiryHours int) *Manager {
	return &Manager{
		secret:      []byte(secret),
		expiryHours: time.Duration(expiryHours) * time.Hour,
	}
}

// Generate creates and signs a new JWT for the given user.
// The token encodes the user's ID, email, role, and verified status.
func (m *Manager) Generate(user *domain.User) (string, error) {
	now := time.Now()

	claims := Claims{
		UserID:     user.ID,
		Email:      user.Email,
		Role:       user.Role,
		IsVerified: user.IsVerified,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.expiryHours)),
			Issuer:    "shikagari-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", err
	}

	return signed, nil
}

// Parse validates a JWT string and returns its claims.
// Returns an error if the token is invalid, expired, or tampered with.
func (m *Manager) Parse(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(t *jwt.Token) (interface{}, error) {
			// Ensure the signing method is what we expect
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return m.secret, nil
		},
	)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
