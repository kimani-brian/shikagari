package interfaces

import (
	"github.com/google/uuid"
	"github.com/shikagari/api/internal/domain"
)

// InquiryRepository defines all database operations for the Inquiry entity.
type InquiryRepository interface {
	// Create persists a new inquiry.
	Create(inquiry *domain.Inquiry) error

	// FindByID retrieves an inquiry by its UUID with all associations preloaded.
	FindByID(id uuid.UUID) (*domain.Inquiry, error)

	// FindByBuyerID returns all inquiries sent by a specific buyer.
	FindByBuyerID(buyerID uuid.UUID, page, perPage int) ([]domain.Inquiry, int64, error)

	// FindBySellerID returns all inquiries received by a specific seller (inbox).
	FindBySellerID(sellerID uuid.UUID, page, perPage int) ([]domain.Inquiry, int64, error)

	// FindByListingID returns all inquiries for a specific listing.
	FindByListingID(listingID uuid.UUID, page, perPage int) ([]domain.Inquiry, int64, error)

	// UpdateReply sets the seller's reply on an inquiry.
	UpdateReply(id uuid.UUID, reply string) error

	// UpdateStatus changes the status of an inquiry.
	UpdateStatus(id uuid.UUID, status domain.InquiryStatus) error

	// BelongsToBuyer checks if an inquiry was sent by the given user.
	BelongsToBuyer(inquiryID, buyerID uuid.UUID) (bool, error)

	// BelongsToSeller checks if an inquiry was addressed to the given seller.
	BelongsToSeller(inquiryID, sellerID uuid.UUID) (bool, error)
}
