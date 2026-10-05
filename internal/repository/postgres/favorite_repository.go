package postgres

import (
	"errors"

	"github.com/google/uuid"
	"github.com/shikagari/api/internal/domain"
	"gorm.io/gorm"
)

type favoriteRepository struct {
	db *gorm.DB
}

// NewFavoriteRepository returns a PostgreSQL-backed FavoriteRepository.
func NewFavoriteRepository(db *gorm.DB) *favoriteRepository {
	return &favoriteRepository{db: db}
}

func (r *favoriteRepository) Add(favorite *domain.Favorite) error {
	return r.db.Create(favorite).Error
}

func (r *favoriteRepository) Remove(userID, listingID uuid.UUID) error {
	return r.db.
		Where("user_id = ? AND listing_id = ?", userID, listingID).
		Delete(&domain.Favorite{}).Error
}

func (r *favoriteRepository) FindByUserID(
	userID uuid.UUID,
	page, perPage int,
) ([]domain.Favorite, int64, error) {
	var favorites []domain.Favorite
	var total int64

	offset := (page - 1) * perPage

	query := r.db.Model(&domain.Favorite{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Listing").
		Preload("Listing.User").
		Preload("Listing.User.DealerProfile").
		Order("created_at DESC").
		Limit(perPage).
		Offset(offset).
		Find(&favorites).Error

	return favorites, total, err
}

func (r *favoriteRepository) IsFavorited(userID, listingID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&domain.Favorite{}).
		Where("user_id = ? AND listing_id = ?", userID, listingID).
		Count(&count).Error
	return count > 0, err
}

func (r *favoriteRepository) FindByUserAndListing(
	userID, listingID uuid.UUID,
) (*domain.Favorite, error) {
	var fav domain.Favorite
	err := r.db.
		Where("user_id = ? AND listing_id = ?", userID, listingID).
		First(&fav).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &fav, nil
}
