package postgres

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shikagari/api/internal/domain"
	"gorm.io/gorm"
)

type sessionRepository struct {
	db *gorm.DB
}

// NewSessionRepository returns a PostgreSQL-backed session store.
func NewSessionRepository(db *gorm.DB) *sessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) Create(session *domain.UserSession) error {
	return r.db.Create(session).Error
}

func (r *sessionRepository) ListByUser(userID uuid.UUID) ([]domain.UserSession, error) {
	var sessions []domain.UserSession
	err := r.db.
		Where("user_id = ?", userID).
		Order("last_active DESC").
		Find(&sessions).Error

	return sessions, err
}

func (r *sessionRepository) FindByID(id uuid.UUID) (*domain.UserSession, error) {
	var session domain.UserSession
	err := r.db.First(&session, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

func (r *sessionRepository) UpdateLastActive(id uuid.UUID, ts time.Time) error {
	updates := map[string]interface{}{
		"last_active": ts,
		"updated_at":  time.Now(),
	}
	return r.db.Model(&domain.UserSession{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *sessionRepository) MarkRevoked(userID uuid.UUID, sessionID uuid.UUID, reason string) error {
	updates := map[string]interface{}{
		"revoked_at":     time.Now(),
		"revoked_reason": reason,
		"updated_at":     time.Now(),
	}

	res := r.db.Model(&domain.UserSession{}).
		Where("id = ? AND user_id = ? AND revoked_at IS NULL", sessionID, userID).
		Updates(updates)

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
