package postgres

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shikagari/api/internal/domain"
	"gorm.io/gorm"
)

type inquiryRepository struct {
	db *gorm.DB
}

// NewInquiryRepository returns a PostgreSQL-backed InquiryRepository.
func NewInquiryRepository(db *gorm.DB) *inquiryRepository {
	return &inquiryRepository{db: db}
}

func (r *inquiryRepository) Create(inquiry *domain.Inquiry) error {
	return r.db.Create(inquiry).Error
}

func (r *inquiryRepository) FindByID(id uuid.UUID) (*domain.Inquiry, error) {
	var inquiry domain.Inquiry
	err := r.db.
		Preload("Listing").
		Preload("Listing.User").
		Preload("Buyer").
		Preload("Seller").
		First(&inquiry, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &inquiry, nil
}

func (r *inquiryRepository) FindByBuyerID(
	buyerID uuid.UUID,
	page, perPage int,
) ([]domain.Inquiry, int64, error) {
	var inquiries []domain.Inquiry
	var total int64

	offset := (page - 1) * perPage
	query := r.db.Model(&domain.Inquiry{}).Where("buyer_id = ?", buyerID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Listing").
		Preload("Listing.User").
		Preload("Seller").
		Order("created_at DESC").
		Limit(perPage).
		Offset(offset).
		Find(&inquiries).Error

	return inquiries, total, err
}

func (r *inquiryRepository) FindBySellerID(
	sellerID uuid.UUID,
	page, perPage int,
) ([]domain.Inquiry, int64, error) {
	var inquiries []domain.Inquiry
	var total int64

	offset := (page - 1) * perPage
	query := r.db.Model(&domain.Inquiry{}).Where("seller_id = ?", sellerID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Listing").
		Preload("Listing.User").
		Preload("Buyer").
		Order("created_at DESC").
		Limit(perPage).
		Offset(offset).
		Find(&inquiries).Error

	return inquiries, total, err
}

func (r *inquiryRepository) FindByListingID(
	listingID uuid.UUID,
	page, perPage int,
) ([]domain.Inquiry, int64, error) {
	var inquiries []domain.Inquiry
	var total int64

	offset := (page - 1) * perPage
	query := r.db.Model(&domain.Inquiry{}).Where("listing_id = ?", listingID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Buyer").
		Order("created_at DESC").
		Limit(perPage).
		Offset(offset).
		Find(&inquiries).Error

	return inquiries, total, err
}

func (r *inquiryRepository) UpdateReply(id uuid.UUID, reply string) error {
	return r.db.Model(&domain.Inquiry{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"reply":      reply,
			"status":     domain.InquiryStatusReplied,
			"updated_at": time.Now(),
		}).Error
}

func (r *inquiryRepository) UpdateStatus(id uuid.UUID, status domain.InquiryStatus) error {
	return r.db.Model(&domain.Inquiry{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error
}

func (r *inquiryRepository) BelongsToBuyer(inquiryID, buyerID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&domain.Inquiry{}).
		Where("id = ? AND buyer_id = ?", inquiryID, buyerID).
		Count(&count).Error
	return count > 0, err
}

func (r *inquiryRepository) BelongsToSeller(inquiryID, sellerID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&domain.Inquiry{}).
		Where("id = ? AND seller_id = ?", inquiryID, sellerID).
		Count(&count).Error
	return count > 0, err
}
