package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PrivateSellerProfile represents an individual (non-business) seller.
// Must be approved by Admin before they can post listings.
type PrivateSellerProfile struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"                 json:"user_id"`

	// Personal details for identity verification
	NationalIDNo    string `gorm:"type:varchar(20);not null;uniqueIndex"          json:"national_id_no"`
	Location        string `gorm:"type:varchar(100);not null"                     json:"location"`
	Address         string `gorm:"type:text"                                      json:"address"`
	ProfilePhotoURL string `gorm:"type:text"                                      json:"profile_photo_url"`
	Bio             string `gorm:"type:text"                                      json:"bio"`

	// Admin approval workflow
	ApprovalStatus ApprovalStatus `gorm:"type:varchar(20);not null;default:'pending'"    json:"approval_status"`
	ApprovedAt     *time.Time     `json:"approved_at"`
	ApprovedByID   *uuid.UUID     `gorm:"type:uuid"                                      json:"approved_by_id"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"                                          json:"-"`

	// Associations
	User       User  `gorm:"foreignKey:UserID"       json:"user,omitempty"`
	ApprovedBy *User `gorm:"foreignKey:ApprovedByID" json:"approved_by,omitempty"`
}
