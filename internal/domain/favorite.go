package domain

import (
	"time"

	"github.com/google/uuid"
)

// Favorite represents a listing saved to a user's wishlist.
// Composite unique index on (UserID, ListingID) prevents duplicates.
type Favorite struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"                       json:"user_id"`
	ListingID uuid.UUID `gorm:"type:uuid;not null;index"                       json:"listing_id"`

	CreatedAt time.Time `json:"created_at"`

	// Associations
	User    User    `gorm:"foreignKey:UserID"    json:"user,omitempty"`
	Listing Listing `gorm:"foreignKey:ListingID" json:"listing,omitempty"`
}
