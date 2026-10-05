package domain

import (
	"strings"
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

// BodyType represents the vehicle's body style/category
type BodyType string

const (
	BodyTypeSUV       BodyType = "SUV"
	BodyTypeSedan     BodyType = "Sedan"
	BodyTypeHatchback BodyType = "Hatchback"
	BodyTypePickup    BodyType = "Pickup"
	BodyTypeCoupe     BodyType = "Coupe"
	BodyTypeEV        BodyType = "EV"
	BodyTypeVan       BodyType = "Van"
	BodyTypeWagon     BodyType = "Wagon"
)

// ListingStatus tracks the visibility/state of a listing
type ListingStatus string

const (
	// ListingPending is a buyer-created listing awaiting NTSA e-logbook
	// review. It is hidden from all public search and detail endpoints
	// until an admin approves it.
	ListingPending  ListingStatus = "pending"
	ListingActive   ListingStatus = "active"
	ListingInactive ListingStatus = "inactive"
	ListingSold     ListingStatus = "sold"
)

// VerificationStatus tracks admin review of a buyer-submitted listing.
// A draft is a buyer listing that has not had its NTSA e-logbook attached yet,
// so it never reaches the admin review queue.
type VerificationStatus string

const (
	VerificationDraft    VerificationStatus = "draft"
	VerificationPending  VerificationStatus = "pending"
	VerificationApproved VerificationStatus = "approved"
	VerificationRejected VerificationStatus = "rejected"
)

// SellerType distinguishes between dealer and private seller listings
type SellerType string

const (
	SellerTypeDealer  SellerType = "dealer"
	SellerTypePrivate SellerType = "private"
)

// IsValidVehicleYear reports whether a model year is plausible:
// between 1980 and next calendar year.
func IsValidVehicleYear(year int) bool {
	now := time.Now().Year()
	return year >= 1980 && year <= now+1
}

// ValidBodyTypes is the allowlist for vehicle body styles.
var ValidBodyTypes = []string{
	string(BodyTypeSUV),
	string(BodyTypeSedan),
	string(BodyTypeHatchback),
	string(BodyTypePickup),
	string(BodyTypeCoupe),
	string(BodyTypeEV),
	string(BodyTypeVan),
	string(BodyTypeWagon),
}

// ParseBodyTypes splits a comma-separated body_type filter (e.g. "Van,SUV")
// and returns only recognised values, so links can cover several styles.
func ParseBodyTypes(raw string) []string {
	var out []string
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		for _, valid := range ValidBodyTypes {
			if part == valid {
				out = append(out, part)
				break
			}
		}
	}
	return out
}

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
	BodyType    BodyType      `gorm:"type:varchar(20);not null;default:'SUV'"        json:"body_type"`
	Status      ListingStatus `gorm:"type:varchar(20);not null;default:'active'"     json:"status"`

	// Vehicle specifications
	Make         string       `gorm:"type:varchar(100);not null;index"               json:"make"`
	Model        string       `gorm:"type:varchar(100);not null;index"               json:"model"`
	Year         int          `gorm:"not null;index"                                 json:"year"`
	Mileage      int          `gorm:"not null"                                       json:"mileage"` // in kilometres
	FuelType     FuelType     `gorm:"type:varchar(20);not null"                      json:"fuel_type"`
	Transmission Transmission `gorm:"type:varchar(20);not null"                      json:"transmission"`
	Drivetrain   string       `gorm:"type:varchar(10)"                               json:"drivetrain"`  // 2WD, 4WD, AWD (optional)
	EngineSize   string       `gorm:"type:varchar(20)"                               json:"engine_size"` // e.g. 3.0L (optional)
	Doors        int          `gorm:"default:0"                                      json:"doors"`       // 0 = not specified
	Color        string       `gorm:"type:varchar(50)"                               json:"color"`

	// Images stored as a PostgreSQL text array (URLs to uploaded images)
	Images pq.StringArray `gorm:"type:text[]"                                    json:"images"`

	// CoverImage is the photo chosen by the seller to represent the listing.
	// Empty means "fall back to the first image".
	CoverImage string `gorm:"type:text" json:"cover_image"`

	// ── Seller verification (buyer-created listings only) ──────────────────
	// Buyers submit identity and ownership proof alongside the listing.
	// The listing stays "pending" and hidden until an admin approves.
	VerificationStatus   VerificationStatus `gorm:"type:varchar(20);not null;default:'pending'" json:"verification_status"`
	VerificationFullName string             `gorm:"type:varchar(150)"                                json:"verification_full_name"`
	VerificationIDNumber string             `gorm:"type:varchar(30)"                                 json:"verification_id_number"`
	VerificationELogbook string             `gorm:"type:text"                                        json:"verification_elogbook_url"`
	VerifiedAt           *time.Time         `json:"verified_at"`
	VerifiedByID         *uuid.UUID         `gorm:"type:uuid"                                        json:"verified_by_id"`
	RejectionReason      string             `gorm:"type:text"                                        json:"rejection_reason"`

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
