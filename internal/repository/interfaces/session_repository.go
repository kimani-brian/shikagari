package interfaces

import (
	"time"

	"github.com/google/uuid"
	"github.com/shikagari/api/internal/domain"
)

// SessionRepository stores login sessions across devices.
type SessionRepository interface {
	Create(session *domain.UserSession) error
	ListByUser(userID uuid.UUID) ([]domain.UserSession, error)
	FindByID(id uuid.UUID) (*domain.UserSession, error)
	UpdateLastActive(id uuid.UUID, ts time.Time) error
	MarkRevoked(userID uuid.UUID, sessionID uuid.UUID, reason string) error
}
