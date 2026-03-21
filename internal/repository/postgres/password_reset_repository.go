package postgres

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shikagari/api/internal/domain"
	"gorm.io/gorm"
)

type passwordResetTokenRepository struct {
	db *gorm.DB
}

// NewPasswordResetTokenRepository returns a repository backed by PostgreSQL.
func NewPasswordResetTokenRepository(db *gorm.DB) *passwordResetTokenRepository {
	return &passwordResetTokenRepository{db: db}
}

func (r *passwordResetTokenRepository) Create(token *domain.PasswordResetToken) error {
	return r.db.Create(token).Error
}

func (r *passwordResetTokenRepository) DeleteByUser(userID uuid.UUID) error {
	return r.db.Where("user_id = ?", userID).Delete(&domain.PasswordResetToken{}).Error
}

func (r *passwordResetTokenRepository) FindByToken(token string) (*domain.PasswordResetToken, error) {
	var reset domain.PasswordResetToken
	err := r.db.Where("token = ?", token).First(&reset).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &reset, nil
}

func (r *passwordResetTokenRepository) MarkUsed(id uuid.UUID, usedAt time.Time) error {
	return r.db.Model(&domain.PasswordResetToken{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"used_at":    usedAt,
			"updated_at": time.Now(),
		}).Error
}

func (r *passwordResetTokenRepository) CleanupExpired(before time.Time) error {
	return r.db.Where("expires_at < ?", before).Delete(&domain.PasswordResetToken{}).Error
}
