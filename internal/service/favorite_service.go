package service

import (
	"errors"

	"github.com/google/uuid"
	"github.com/shikagari/api/internal/domain"
	"github.com/shikagari/api/internal/dto"
	"github.com/shikagari/api/internal/repository/interfaces"
)

// FavoriteService handles saving and removing listings from a user's wishlist.
type FavoriteService struct {
	favoriteRepo interfaces.FavoriteRepository
	listingRepo  interfaces.ListingRepository
}

// NewFavoriteService constructs a FavoriteService with its dependencies.
func NewFavoriteService(
	favoriteRepo interfaces.FavoriteRepository,
	listingRepo interfaces.ListingRepository,
) *FavoriteService {
	return &FavoriteService{
		favoriteRepo: favoriteRepo,
		listingRepo:  listingRepo,
	}
}

// Toggle adds the listing to the user's favorites if not already saved,
// or removes it if it is. Returns the resulting state.
func (s *FavoriteService) Toggle(
	userID uuid.UUID,
	listingID uuid.UUID,
) (*dto.FavoriteToggleResponse, error) {
	// ── 1. Verify the listing exists and is active ────────────────────────────
	listing, err := s.listingRepo.FindByID(listingID)
	if err != nil {
		return nil, errors.New("failed to retrieve listing")
	}
	if listing == nil {
		return nil, errors.New("listing not found")
	}
	if listing.Status != domain.ListingActive {
		return nil, errors.New("this listing is no longer active")
	}

	// ── 2. Check current state ────────────────────────────────────────────────
	isFavorited, err := s.favoriteRepo.IsFavorited(userID, listingID)
	if err != nil {
		return nil, errors.New("failed to check favorite status")
	}

	if isFavorited {
		// Remove from favorites
		if err := s.favoriteRepo.Remove(userID, listingID); err != nil {
			return nil, errors.New("failed to remove listing from favorites")
		}
		return &dto.FavoriteToggleResponse{
			ListingID: listingID,
			Saved:     false,
			Message:   "Listing removed from your favorites",
		}, nil
	}

	// Add to favorites
	favorite := &domain.Favorite{
		UserID:    userID,
		ListingID: listingID,
	}
	if err := s.favoriteRepo.Add(favorite); err != nil {
		return nil, errors.New("failed to add listing to favorites")
	}

	return &dto.FavoriteToggleResponse{
		ListingID: listingID,
		Saved:     true,
		Message:   "Listing saved to your favorites",
	}, nil
}

// GetMyFavorites returns a paginated list of the authenticated user's saved listings.
func (s *FavoriteService) GetMyFavorites(
	userID uuid.UUID,
	page, perPage int,
) ([]dto.FavoriteResponse, int64, error) {
	favorites, total, err := s.favoriteRepo.FindByUserID(userID, page, perPage)
	if err != nil {
		return nil, 0, errors.New("failed to retrieve favorites")
	}

	var result []dto.FavoriteResponse
	for _, fav := range favorites {
		result = append(result, dto.FavoriteResponse{
			ID:        fav.ID,
			UserID:    fav.UserID,
			ListingID: fav.ListingID,
			Listing:   dto.ToListingCardResponse(fav.Listing),
			CreatedAt: fav.CreatedAt,
		})
	}

	return result, total, nil
}

// IsFavorited checks whether a specific listing is in the user's favorites.
// Used to enrich listing detail responses with the saved state.
func (s *FavoriteService) IsFavorited(userID, listingID uuid.UUID) (bool, error) {
	return s.favoriteRepo.IsFavorited(userID, listingID)
}
