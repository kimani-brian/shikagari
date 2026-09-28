package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ApprovalStatus tracks admin review state for seller profiles
type ApprovalStatus string

const (
	ApprovalPending  ApprovalStatus = "pending"
	ApprovalApproved ApprovalStatus = "approved"
	ApprovalRejected ApprovalStatus = "rejected"
)

// DealerProfile represents a business/dealership account.
// Must be approved by Admin before the dealer can post listings.
type DealerProfile struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"                 json:"user_id"`

	// Business information
	BusinessName  string `gorm:"type:varchar(200);not null"                     json:"business_name"`
	BusinessRegNo string `gorm:"type:varchar(100);index"                        json:"business_reg_no"`
	Location      string `gorm:"type:varchar(100);not null"                     json:"location"`
	Address       string `gorm:"type:text"                                      json:"address"`
	LogoURL       string `gorm:"type:text"                                      json:"logo_url"`
	Description   string `gorm:"type:text"                                      json:"description"`

	// KRA PIN for tax compliance (Kenya-specific)
	KRAPIN string `gorm:"type:varchar(20)"                               json:"kra_pin"`

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
