package interfaces

import (
	"github.com/google/uuid"
	"github.com/shikagari/api/internal/domain"
)

// SecurityEventRepository persists account security activity entries.
type SecurityEventRepository interface {
	Create(event *domain.SecurityEvent) error
	ListRecentByUser(userID uuid.UUID, limit int) ([]domain.SecurityEvent, error)
}
