package dto

import (
	"time"

	"github.com/google/uuid"
)

// ── Response DTOs ─────────────────────────────────────────────────────────────

// FavoriteResponse is returned when a listing is saved to a wishlist.
type FavoriteResponse struct {
	ID        uuid.UUID           `json:"id"`
	UserID    uuid.UUID           `json:"user_id"`
	ListingID uuid.UUID           `json:"listing_id"`
	Listing   ListingCardResponse `json:"listing"`
	CreatedAt time.Time           `json:"created_at"`
}

// FavoriteToggleResponse is returned by the toggle endpoint
// to indicate whether the listing was added or removed.
type FavoriteToggleResponse struct {
	ListingID uuid.UUID `json:"listing_id"`
	Saved     bool      `json:"saved"` // true = added, false = removed
	Message   string    `json:"message"`
}
