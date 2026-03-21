package interfaces

import (
	"time"

	"github.com/google/uuid"
	"github.com/shikagari/api/internal/domain"
)

// PasswordResetTokenRepository handles persistence for password reset tokens.
type PasswordResetTokenRepository interface {
	Create(token *domain.PasswordResetToken) error
	DeleteByUser(userID uuid.UUID) error
	FindByToken(token string) (*domain.PasswordResetToken, error)
	MarkUsed(id uuid.UUID, usedAt time.Time) error
	CleanupExpired(before time.Time) error
}
