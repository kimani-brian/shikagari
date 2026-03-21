package postgres

import (
	"github.com/google/uuid"
	"github.com/shikagari/api/internal/domain"
	"gorm.io/gorm"
)

type securityEventRepository struct {
	db *gorm.DB
}

// NewSecurityEventRepository provides a PostgreSQL-backed implementation.
func NewSecurityEventRepository(db *gorm.DB) *securityEventRepository {
	return &securityEventRepository{db: db}
}

func (r *securityEventRepository) Create(event *domain.SecurityEvent) error {
	return r.db.Create(event).Error
}

func (r *securityEventRepository) ListRecentByUser(userID uuid.UUID, limit int) ([]domain.SecurityEvent, error) {
	if limit <= 0 {
		limit = 20
	}

	var events []domain.SecurityEvent
	err := r.db.
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&events).Error

	return events, err
}
