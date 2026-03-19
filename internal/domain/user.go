package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRole defines the access level of a registered user
type UserRole string

const (
	RoleBuyer  UserRole = "buyer"
	RoleSeller UserRole = "seller"
	RoleAdmin  UserRole = "admin"
)

// User represents a registered platform user.
// Buyers can browse and inquire; sellers can list vehicles after profile approval.
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	FullName     string    `gorm:"type:varchar(150);not null"                    json:"full_name"`
	Email        string    `gorm:"type:varchar(255);uniqueIndex;not null"         json:"email"`
	Phone        string    `gorm:"type:varchar(20);not null"                      json:"phone"`
	PasswordHash string    `gorm:"type:text;not null"                             json:"-"`
	Role         UserRole  `gorm:"type:varchar(20);not null;default:'buyer'"      json:"role"`

	// IsVerified is the "Verified Badge" — set by Admin for trusted sellers
	IsVerified bool `gorm:"not null;default:false"                         json:"is_verified"`

	// IsActive allows soft-disabling of accounts by Admin
	IsActive bool `gorm:"not null;default:true"                          json:"is_active"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"                                          json:"-"`

	// Associations
	DealerProfile        *DealerProfile        `gorm:"foreignKey:UserID"  json:"dealer_profile,omitempty"`
	PrivateSellerProfile *PrivateSellerProfile `gorm:"foreignKey:UserID" json:"private_seller_profile,omitempty"`
	Listings             []Listing             `gorm:"foreignKey:UserID"  json:"-"`
	Favorites            []Favorite            `gorm:"foreignKey:UserID"  json:"-"`
	Inquiries            []Inquiry             `gorm:"foreignKey:BuyerID" json:"-"`
}
