package postgres

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shikagari/api/internal/domain"
	"github.com/shikagari/api/internal/dto"
	"gorm.io/gorm"
)

type listingRepository struct {
	db *gorm.DB
}

// NewListingRepository returns a PostgreSQL-backed ListingRepository.
func NewListingRepository(db *gorm.DB) *listingRepository {
	return &listingRepository{db: db}
}

func (r *listingRepository) Create(listing *domain.Listing) error {
	return r.db.Create(listing).Error
}

func (r *listingRepository) FindByID(id uuid.UUID) (*domain.Listing, error) {
	var listing domain.Listing
	err := r.db.
		Preload("User").
		Preload("User.DealerProfile").
		Preload("User.PrivateSellerProfile").
		First(&listing, "listings.id = ? AND listings.deleted_at IS NULL", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &listing, nil
}

func (r *listingRepository) Update(listing *domain.Listing) error {
	return r.db.Save(listing).Error
}

func (r *listingRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.Listing{}, "id = ?", id).Error
}

// Search builds a dynamic query from the filter request and returns
// a paginated slice of active listings.
func (r *listingRepository) Search(filters dto.ListingFilterRequest) ([]domain.Listing, int64, error) {
	var listings []domain.Listing
	var total int64

	// Cap per_page at 50 to protect against large payload requests
	if filters.PerPage > 50 {
		filters.PerPage = 50
	}
	offset := (filters.Page - 1) * filters.PerPage

	query := r.db.Model(&domain.Listing{}).
		Where("listings.deleted_at IS NULL").
		Where("listings.status = ?", domain.ListingActive)

	// ── Filters ───────────────────────────────────────────────────────────────

	if filters.Location != "" {
		query = query.Where("LOWER(listings.location) = LOWER(?)", filters.Location)
	}
	if filters.BodyType != "" {
		query = query.Where("listings.body_type = ?", filters.BodyType)
	}
	if filters.Make != "" {
		query = query.Where("LOWER(listings.make) = LOWER(?)", filters.Make)
	}
	if filters.Model != "" {
		query = query.Where("LOWER(listings.model) = LOWER(?)", filters.Model)
	}
	if filters.FuelType != "" {
		query = query.Where("listings.fuel_type = ?", filters.FuelType)
	}
	if filters.Transmission != "" {
		query = query.Where("listings.transmission = ?", filters.Transmission)
	}
	if filters.SellerType != "" {
		query = query.Where("listings.seller_type = ?", filters.SellerType)
	}
	if filters.MinPriceKES > 0 {
		query = query.Where("listings.price_kes >= ?", filters.MinPriceKES)
	}
	if filters.MaxPriceKES > 0 {
		query = query.Where("listings.price_kes <= ?", filters.MaxPriceKES)
	}
	if filters.MinYear > 0 {
		query = query.Where("listings.year >= ?", filters.MinYear)
	}
	if filters.MaxYear > 0 {
		query = query.Where("listings.year <= ?", filters.MaxYear)
	}

		if filters.UserID != "" {
		if uid, err := uuid.Parse(filters.UserID); err == nil {
			query = query.Where("listings.user_id = ?", uid)
		}
	}
	if filters.DealerID != "" {
		if did, err := uuid.Parse(filters.DealerID); err == nil {
			// Resolve dealer profile → user_id then filter
			var dealerUserID uuid.UUID
			if err := r.db.Raw("SELECT user_id FROM dealer_profiles WHERE id = ? AND deleted_at IS NULL", did).Scan(&dealerUserID).Error; err == nil && dealerUserID != uuid.Nil {
				query = query.Where("listings.user_id = ?", dealerUserID)
			} else {
				// Invalid dealer_id yields no results
				query = query.Where("1 = 0")
			}
		}
	}

	// ── Full-text keyword search across title and description ─────────────────
	if filters.Search != "" {
		searchTerm := "%" + strings.ToLower(filters.Search) + "%"
		query = query.Where(
			"LOWER(listings.title) LIKE ? OR LOWER(listings.description) LIKE ?",
			searchTerm, searchTerm,
		)
	}

	// ── Count before pagination ───────────────────────────────────────────────
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// ── Sorting ───────────────────────────────────────────────────────────────
	orderClause := buildOrderClause(filters.SortBy)

	err := query.
		Preload("User").
		Preload("User.DealerProfile").
		Preload("User.PrivateSellerProfile").
		Order(orderClause).
		Limit(filters.PerPage).
		Offset(offset).
		Find(&listings).Error

	return listings, total, err
}

func (r *listingRepository) FindByUserID(
	userID uuid.UUID,
	page, perPage int,
) ([]domain.Listing, int64, error) {
	var listings []domain.Listing
	var total int64

	offset := (page - 1) * perPage

	query := r.db.Model(&domain.Listing{}).
		Where("user_id = ? AND deleted_at IS NULL", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("User").
		Order("created_at DESC").
		Limit(perPage).
		Offset(offset).
		Find(&listings).Error

	return listings, total, err
}

func (r *listingRepository) UpdateImages(id uuid.UUID, images []string) error {
	return r.db.Model(&domain.Listing{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"images":     images,
			"updated_at": time.Now(),
		}).Error
}

func (r *listingRepository) IncrementViewCount(id uuid.UUID) error {
	// Use a raw UPDATE for atomicity — avoids race conditions on concurrent views
	return r.db.Exec(
		"UPDATE listings SET view_count = view_count + 1, updated_at = ? WHERE id = ?",
		time.Now(), id,
	).Error
}

func (r *listingRepository) UpdateStatus(id uuid.UUID, status domain.ListingStatus) error {
	return r.db.Model(&domain.Listing{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error
}

func (r *listingRepository) BelongsToUser(listingID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&domain.Listing{}).
		Where("id = ? AND user_id = ? AND deleted_at IS NULL", listingID, userID).
		Count(&count).Error
	return count > 0, err
}

// buildOrderClause maps a sort_by query param to a safe SQL ORDER BY clause.
func buildOrderClause(sortBy string) string {
	orderMap := map[string]string{
		"price_asc":  "listings.price_kes ASC",
		"price_desc": "listings.price_kes DESC",
		"year_asc":   "listings.year ASC",
		"year_desc":  "listings.year DESC",
		"newest":     "listings.created_at DESC",
	}

	if clause, ok := orderMap[sortBy]; ok {
		return clause
	}

	// Default: newest first
	return fmt.Sprintf("listings.created_at DESC")
}
