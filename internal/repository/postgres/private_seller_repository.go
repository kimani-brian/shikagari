package postgres

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shikagari/api/internal/domain"
	"gorm.io/gorm"
)

type privateSellerRepository struct {
	db *gorm.DB
}

// NewPrivateSellerRepository returns a PostgreSQL-backed PrivateSellerRepository.
func NewPrivateSellerRepository(db *gorm.DB) *privateSellerRepository {
	return &privateSellerRepository{db: db}
}

func (r *privateSellerRepository) Create(profile *domain.PrivateSellerProfile) error {
	return r.db.Create(profile).Error
}

func (r *privateSellerRepository) FindByID(id uuid.UUID) (*domain.PrivateSellerProfile, error) {
	var profile domain.PrivateSellerProfile
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

func (r *privateSellerRepository) FindByUserID(userID uuid.UUID) (*domain.PrivateSellerProfile, error) {
	var profile domain.PrivateSellerProfile
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

func (r *privateSellerRepository) Update(profile *domain.PrivateSellerProfile) error {
	return r.db.Save(profile).Error
}

func (r *privateSellerRepository) UpdateApprovalStatus(
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
	return r.db.Model(&domain.PrivateSellerProfile{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *privateSellerRepository) UpdateProfilePhotoURL(id uuid.UUID, photoURL string) error {
	return r.db.Model(&domain.PrivateSellerProfile{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"profile_photo_url": photoURL,
			"updated_at":        time.Now(),
		}).Error
}

func (r *privateSellerRepository) List(
	status domain.ApprovalStatus,
	page, perPage int,
) ([]domain.PrivateSellerProfile, int64, error) {
	var profiles []domain.PrivateSellerProfile
	var total int64

	offset := (page - 1) * perPage
	query := r.db.Model(&domain.PrivateSellerProfile{}).Where("deleted_at IS NULL")

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

func (r *privateSellerRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.PrivateSellerProfile{}, "id = ?", id).Error
}

func (r *privateSellerRepository) ExistsByUserID(userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&domain.PrivateSellerProfile{}).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Count(&count).Error
	return count > 0, err
}

func (r *privateSellerRepository) ExistsByNationalID(nationalID string) (bool, error) {
	var count int64
	err := r.db.Model(&domain.PrivateSellerProfile{}).
		Where("national_id_no = ? AND deleted_at IS NULL", nationalID).
		Count(&count).Error
	return count > 0, err
}
