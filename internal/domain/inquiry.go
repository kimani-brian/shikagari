package domain

import (
	"time"

	"github.com/google/uuid"
)

// InquiryStatus tracks whether a seller has responded to an inquiry
type InquiryStatus string

const (
	InquiryStatusOpen    InquiryStatus = "open"
	InquiryStatusReplied InquiryStatus = "replied"
	InquiryStatusClosed  InquiryStatus = "closed"
)

// Inquiry represents a message sent by a buyer to a seller about a listing.
// Authentication is required to send an inquiry.
type Inquiry struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ListingID uuid.UUID `gorm:"type:uuid;not null;index"                       json:"listing_id"`
	BuyerID   uuid.UUID `gorm:"type:uuid;not null;index"                       json:"buyer_id"`
	SellerID  uuid.UUID `gorm:"type:uuid;not null;index"                       json:"seller_id"`

	// The buyer's message
	Message string `gorm:"type:text;not null"                             json:"message"`

	// Optional seller reply stored inline (keeps MVP simple; extend to threads later)
	Reply string `gorm:"type:text"                                      json:"reply"`

	Status InquiryStatus `gorm:"type:varchar(20);not null;default:'open'"       json:"status"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Associations
	Listing Listing `gorm:"foreignKey:ListingID" json:"listing,omitempty"`
	Buyer   User    `gorm:"foreignKey:BuyerID"   json:"buyer,omitempty"`
	Seller  User    `gorm:"foreignKey:SellerID"  json:"seller,omitempty"`
}
