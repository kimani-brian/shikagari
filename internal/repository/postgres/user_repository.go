package postgres

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shikagari/api/internal/domain"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository returns a PostgreSQL-backed UserRepository.
func NewUserRepository(db *gorm.DB) *userRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *domain.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindByID(id uuid.UUID) (*domain.User, error) {
	var user domain.User
	err := r.db.
		Preload("DealerProfile").
		First(&user, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // caller decides how to handle not-found
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(email string) (*domain.User, error) {
	var user domain.User
	err := r.db.
		Preload("DealerProfile").
		First(&user, "email = ? AND deleted_at IS NULL", email).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(user *domain.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) UpdatePassword(id uuid.UUID, passwordHash string) error {
	return r.db.Model(&domain.User{}).
		Where("id = ?", id).
		Update("password_hash", passwordHash).Error
}

func (r *userRepository) SetVerified(id uuid.UUID, verified bool) error {
	return r.db.Model(&domain.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_verified": verified,
			"updated_at":  time.Now(),
		}).Error
}

func (r *userRepository) SetActive(id uuid.UUID, active bool) error {
	return r.db.Model(&domain.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_active":  active,
			"updated_at": time.Now(),
		}).Error
}

func (r *userRepository) SetRole(id uuid.UUID, role domain.UserRole) error {
	return r.db.Model(&domain.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"role":       role,
			"updated_at": time.Now(),
		}).Error
}

func (r *userRepository) List(page, perPage int) ([]domain.User, int64, error) {
	var users []domain.User
	var total int64

	offset := (page - 1) * perPage

	if err := r.db.Model(&domain.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.
		Preload("DealerProfile").
		Order("created_at DESC").
		Limit(perPage).
		Offset(offset).
		Find(&users).Error

	return users, total, err
}

func (r *userRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.User{}, "id = ?", id).Error
}

func (r *userRepository) ExistsByEmail(email string) (bool, error) {
	var count int64
	err := r.db.Model(&domain.User{}).
		Where("email = ? AND deleted_at IS NULL", email).
		Count(&count).Error
	return count > 0, err
}
