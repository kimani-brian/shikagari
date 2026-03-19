package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// FuelType defines the vehicle's fuel type
type FuelType string

const (
	FuelPetrol   FuelType = "petrol"
	FuelDiesel   FuelType = "diesel"
	FuelHybrid   FuelType = "hybrid"
	FuelElectric FuelType = "electric"
)

// Transmission defines the gearbox type
type Transmission string

const (
	TransmissionAutomatic Transmission = "automatic"
	TransmissionManual    Transmission = "manual"
)

// ListingStatus tracks the visibility/state of a listing
type ListingStatus string

const (
	ListingActive   ListingStatus = "active"
	ListingInactive ListingStatus = "inactive"
	ListingSold     ListingStatus = "sold"
)

// SellerType distinguishes between dealer and private seller listings
type SellerType string

const (
	SellerTypeDealer  SellerType = "dealer"
	SellerTypePrivate SellerType = "private"
)

// Listing represents a vehicle listed for sale on ShikaGari.
// Prices are stored in Kenya Shillings (KES).
type Listing struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID     uuid.UUID  `gorm:"type:uuid;not null;index"                       json:"user_id"`
	SellerType SellerType `gorm:"type:varchar(20);not null"                      json:"seller_type"`

	// Core listing details
	Title       string        `gorm:"type:varchar(255);not null"                     json:"title"`
	Description string        `gorm:"type:text"                                      json:"description"`
	PriceKES    float64       `gorm:"type:numeric(15,2);not null"                    json:"price_kes"` // Kenya Shillings
	Location    string        `gorm:"type:varchar(100);not null;index"               json:"location"`
	Status      ListingStatus `gorm:"type:varchar(20);not null;default:'active'"     json:"status"`

	// Vehicle specifications
	Make         string       `gorm:"type:varchar(100);not null;index"               json:"make"`
	Model        string       `gorm:"type:varchar(100);not null;index"               json:"model"`
	Year         int          `gorm:"not null;index"                                 json:"year"`
	Mileage      int          `gorm:"not null"                                       json:"mileage"` // in kilometres
	FuelType     FuelType     `gorm:"type:varchar(20);not null"                      json:"fuel_type"`
	Transmission Transmission `gorm:"type:varchar(20);not null"                      json:"transmission"`
	Color        string       `gorm:"type:varchar(50)"                               json:"color"`

	// Images stored as a PostgreSQL text array (URLs to uploaded images)
	Images pq.StringArray `gorm:"type:text[]"                                    json:"images"`

	// View counter — lightweight engagement metric
	ViewCount int `gorm:"not null;default:0"                             json:"view_count"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"                                          json:"-"`

	// Associations
	User      User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Favorites []Favorite `gorm:"foreignKey:ListingID" json:"-"`
	Inquiries []Inquiry  `gorm:"foreignKey:ListingID" json:"-"`
}
