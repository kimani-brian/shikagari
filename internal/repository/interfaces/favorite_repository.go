package interfaces

import (
	"github.com/google/uuid"
	"github.com/shikagari/api/internal/domain"
)

// FavoriteRepository defines all database operations for the Favorite entity.
type FavoriteRepository interface {
	// Add saves a listing to a user's favorites.
	Add(favorite *domain.Favorite) error

	// Remove deletes a saved listing from a user's favorites.
	Remove(userID, listingID uuid.UUID) error

	// FindByUserID returns all favorites for a given user with listing preloaded.
	FindByUserID(userID uuid.UUID, page, perPage int) ([]domain.Favorite, int64, error)

	// IsFavorited checks if a user has already saved a specific listing.
	IsFavorited(userID, listingID uuid.UUID) (bool, error)

	// FindByUserAndListing retrieves a specific favorite record.
	FindByUserAndListing(userID, listingID uuid.UUID) (*domain.Favorite, error)
}
