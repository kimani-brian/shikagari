package interfaces

import (
	"github.com/google/uuid"
	"github.com/shikagari/api/internal/domain"
	"github.com/shikagari/api/internal/dto"
)

// ListingRepository defines all database operations for the Listing entity.
type ListingRepository interface {
	// Create persists a new vehicle listing.
	Create(listing *domain.Listing) error

	// FindByID retrieves a listing by its UUID with all associations preloaded.
	FindByID(id uuid.UUID) (*domain.Listing, error)

	// Update persists changes to an existing listing.
	Update(listing *domain.Listing) error

	// Delete soft-deletes a listing.
	Delete(id uuid.UUID) error

	// Search returns a paginated, filtered list of active listings.
	Search(filters dto.ListingFilterRequest) ([]domain.Listing, int64, error)

	// FindByUserID returns all listings created by a specific seller.
	FindByUserID(userID uuid.UUID, page, perPage int) ([]domain.Listing, int64, error)

	// UpdateImages replaces the images array for a listing.
	UpdateImages(id uuid.UUID, images []string) error

	// FindByVerificationStatus returns listings awaiting admin review,
	// newest first, for the seller-verification queue.
	FindByVerificationStatus(status domain.VerificationStatus, page, perPage int) ([]domain.Listing, int64, error)

	// IncrementViewCount atomically increments the view counter for a listing.
	IncrementViewCount(id uuid.UUID) error

	// UpdateStatus changes the status of a listing (active/inactive/sold).
	UpdateStatus(id uuid.UUID, status domain.ListingStatus) error

	// BelongsToUser checks if a listing is owned by the given user.
	BelongsToUser(listingID, userID uuid.UUID) (bool, error)
}
