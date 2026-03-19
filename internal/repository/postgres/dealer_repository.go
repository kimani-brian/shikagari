package postgres

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shikagari/api/internal/domain"
	"gorm.io/gorm"
)

type dealerRepository struct {
	db *gorm.DB
}

// NewDealerRepository returns a PostgreSQL-backed DealerRepository.
func NewDealerRepository(db *gorm.DB) *dealerRepository {
	return &dealerRepository{db: db}
}

func (r *dealerRepository) Create(profile *domain.DealerProfile) error {
	return r.db.Create(profile).Error
}

func (r *dealerRepository) FindByID(id uuid.UUID) (*domain.DealerProfile, error) {
	var profile domain.DealerProfile
	err := r.db.
		Preload("User").
		Preload("ApprovedBy").
		First(&profile, "id = ? AND deleted_at IS NULL", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &profile, nil
}

func (r *dealerRepository) FindByUserID(userID uuid.UUID) (*domain.DealerProfile, error) {
	var profile domain.DealerProfile
	err := r.db.
		Preload("User").
		First(&profile, "user_id = ? AND deleted_at IS NULL", userID).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &profile, nil
}

func (r *dealerRepository) Update(profile *domain.DealerProfile) error {
	return r.db.Save(profile).Error
}

func (r *dealerRepository) UpdateApprovalStatus(
	id uuid.UUID,
	status domain.ApprovalStatus,
	approvedByID uuid.UUID,
) error {
	now := time.Now()
	updates := map[string]interface{}{
		"approval_status": status,
		"approved_by_id":  approvedByID,
		"updated_at":      now,
	}
	if status == domain.ApprovalApproved {
		updates["approved_at"] = now
	}
	return r.db.Model(&domain.DealerProfile{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *dealerRepository) UpdateLogoURL(id uuid.UUID, logoURL string) error {
	return r.db.Model(&domain.DealerProfile{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"logo_url":   logoURL,
			"updated_at": time.Now(),
		}).Error
}

func (r *dealerRepository) List(
	status domain.ApprovalStatus,
	page, perPage int,
) ([]domain.DealerProfile, int64, error) {
	var profiles []domain.DealerProfile
	var total int64

	offset := (page - 1) * perPage
	query := r.db.Model(&domain.DealerProfile{}).Where("deleted_at IS NULL")

	if status != "" {
		query = query.Where("approval_status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("User").
		Order("created_at DESC").
		Limit(perPage).
		Offset(offset).
		Find(&profiles).Error

	return profiles, total, err
}

func (r *dealerRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.DealerProfile{}, "id = ?", id).Error
}

func (r *dealerRepository) ExistsByUserID(userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&domain.DealerProfile{}).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Count(&count).Error
	return count > 0, err
}

func (r *dealerRepository) ExistsByBusinessRegNo(regNo string) (bool, error) {
	var count int64
	err := r.db.Model(&domain.DealerProfile{}).
		Where("business_reg_no = ? AND deleted_at IS NULL", regNo).
		Count(&count).Error
	return count > 0, err
}
