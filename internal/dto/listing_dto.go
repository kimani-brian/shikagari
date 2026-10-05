package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shikagari/api/internal/domain"
)

// ── Request DTOs ─────────────────────────────────────────────────────────────

// CreateListingRequest is the payload for POST /listings.
//
// Buyers must include the verification fields (full name, ID number and an
// NTSA e-logbook upload reference). Their listing is created as "pending"
// and stays hidden from public search until an admin approves it.
// Dealers list straight away once their dealer profile is approved.
type CreateListingRequest struct {
	Title        string  `json:"title"        binding:"required,min=5,max=255"`
	Description  string  `json:"description"  binding:"omitempty,max=5000"`
	PriceKES     float64 `json:"price_kes"    binding:"required,gte=0"`
	Location     string  `json:"location"     binding:"required,kenyacounty"`
	BodyType     string  `json:"body_type"    binding:"required,oneof=SUV Sedan Hatchback Pickup Coupe EV Van Wagon"`
	Make         string  `json:"make"         binding:"required,min=1,max=100"`
	Model        string  `json:"model"        binding:"required,min=1,max=100"`
	Year         int     `json:"year"         binding:"required,vehicleyear"`
	Mileage      int     `json:"mileage"      binding:"required,gte=0"`
	FuelType     string  `json:"fuel_type"    binding:"required,oneof=petrol diesel hybrid electric"`
	Transmission string  `json:"transmission" binding:"required,oneof=automatic manual"`
	Drivetrain   string  `json:"drivetrain"   binding:"omitempty,oneof=2WD 4WD AWD"`
	EngineSize   string  `json:"engine_size"  binding:"omitempty,max=20"`
	Doors        *int    `json:"doors"        binding:"omitempty,gte=2,lte=6"`
	Color        string  `json:"color"        binding:"omitempty,max=50"`

	// ── Seller verification (required for buyers) ──────────────────────────
	VerificationFullName string `json:"verification_full_name" binding:"omitempty,max=150"`
	VerificationIDNumber string `json:"verification_id_number" binding:"omitempty,max=30"`
	VerificationELogbook string `json:"verification_elogbook_url" binding:"omitempty,max=500"`

	// Images are uploaded separately via POST /listings/:id/images
	// and are not part of the initial create payload.
}

// AdminReviewListingRequest is the payload for approving or rejecting a
// pending buyer listing.
type AdminReviewListingRequest struct {
	VerificationStatus string `json:"verification_status" binding:"required,oneof=approved rejected"`
	RejectionReason    string `json:"rejection_reason"   binding:"omitempty,max=500"`
}

// UpdateListingRequest is the payload for PATCH /listings/:id
// All fields are optional — only provided fields are updated.
type UpdateListingRequest struct {
	Title        *string  `json:"title"        binding:"omitempty,min=5,max=255"`
	Description  *string  `json:"description"  binding:"omitempty,max=5000"`
	PriceKES     *float64 `json:"price_kes"    binding:"omitempty,gte=0"`
	Location     *string  `json:"location"     binding:"omitempty,kenyacounty"`
	BodyType     *string  `json:"body_type"    binding:"omitempty,oneof=SUV Sedan Hatchback Pickup Coupe EV Van Wagon"`
	Make         *string  `json:"make"         binding:"omitempty,min=1,max=100"`
	Model        *string  `json:"model"        binding:"omitempty,min=1,max=100"`
	Year         *int     `json:"year"         binding:"omitempty,vehicleyear"`
	Mileage      *int     `json:"mileage"      binding:"omitempty,gte=0"`
	FuelType     *string  `json:"fuel_type"    binding:"omitempty,oneof=petrol diesel hybrid electric"`
	Transmission *string  `json:"transmission" binding:"omitempty,oneof=automatic manual"`
	Drivetrain   *string  `json:"drivetrain"   binding:"omitempty,oneof=2WD 4WD AWD"`
	EngineSize   *string  `json:"engine_size"  binding:"omitempty,max=20"`
	Doors        *int     `json:"doors"        binding:"omitempty,gte=2,lte=6"`
	Color        *string  `json:"color"        binding:"omitempty,max=50"`
	Status       *string  `json:"status"       binding:"omitempty,oneof=pending active inactive sold"`
}

// ListingFilterRequest maps query parameters for GET /listings.
// body_type accepts a single style or a comma-separated list (e.g. "Van,SUV").
type ListingFilterRequest struct {
	Search       string  `form:"search"`
	Location     string  `form:"location"`
	BodyType     string  `form:"body_type"`
	Make         string  `form:"make"`
	Model        string  `form:"model"`
	MinYear      int     `form:"min_year"`
	MaxYear      int     `form:"max_year"`
	MinPriceKES  float64 `form:"min_price"`
	MaxPriceKES  float64 `form:"max_price"`
	FuelType     string  `form:"fuel_type"    binding:"omitempty,oneof=petrol diesel hybrid electric"`
	Transmission string  `form:"transmission" binding:"omitempty,oneof=automatic manual"`
	Drivetrain   string  `form:"drivetrain"   binding:"omitempty,oneof=2WD 4WD AWD"`
	Doors        int     `form:"doors"`
	SellerType   string  `form:"seller_type"  binding:"omitempty,oneof=dealer private"`
	DealerID     string  `form:"dealer_id"` // dealer profile UUID — filters to that dealer's inventory
	UserID       string  `form:"user_id"`   // user UUID — filters to that user's listings
	SortBy       string  `form:"sort_by"      binding:"omitempty,oneof=price_asc price_desc year_asc year_desc newest"`
	Page         int     `form:"page,default=1"`
	PerPage      int     `form:"per_page,default=20"`
}

// ── Response DTOs ─────────────────────────────────────────────────────────────

// ListingResponse is the full listing shape returned by detail endpoints.
type ListingResponse struct {
	ID           uuid.UUID            `json:"id"`
	Title        string               `json:"title"`
	Description  string               `json:"description"`
	PriceKES     float64              `json:"price_kes"`
	Location     string               `json:"location"`
	BodyType     domain.BodyType      `json:"body_type"`
	Status       domain.ListingStatus `json:"status"`
	SellerType   domain.SellerType    `json:"seller_type"`
	Make         string               `json:"make"`
	Model        string               `json:"model"`
	Year         int                  `json:"year"`
	Mileage      int                  `json:"mileage"`
	FuelType     domain.FuelType      `json:"fuel_type"`
	Transmission domain.Transmission  `json:"transmission"`
	Drivetrain   string               `json:"drivetrain"`
	EngineSize   string               `json:"engine_size"`
	Doors        int                  `json:"doors"`
	Color        string               `json:"color"`
	Images       []string             `json:"images"`
	CoverImage   string               `json:"cover_image"` // seller-chosen card photo
	ViewCount    int                  `json:"view_count"`
	Seller       UserSummary          `json:"seller"`

	// Seller verification — present on buyer-created listings
	VerificationStatus   domain.VerificationStatus `json:"verification_status"`
	VerificationFullName string                    `json:"verification_full_name,omitempty"`
	VerificationIDNumber string                    `json:"verification_id_number,omitempty"`
	VerificationELogbook string                    `json:"verification_elogbook_url,omitempty"`
	VerifiedAt           *time.Time                `json:"verified_at,omitempty"`
	RejectionReason      string                    `json:"rejection_reason,omitempty"`

	// Dealer profile summary — only set on dealer listings
	DealerProfile *DealerSummary `json:"dealer_profile,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ListingCardResponse is the lightweight shape used in list/search results.
// Excludes description and full seller detail to keep payloads small —
// critical for mobile and low-bandwidth users.
type ListingCardResponse struct {
	ID           uuid.UUID            `json:"id"`
	Title        string               `json:"title"`
	PriceKES     float64              `json:"price_kes"`
	Location     string               `json:"location"`
	BodyType     domain.BodyType      `json:"body_type"`
	Status       domain.ListingStatus `json:"status"`
	Make         string               `json:"make"`
	Model        string               `json:"model"`
	Year         int                  `json:"year"`
	Mileage      int                  `json:"mileage"`
	FuelType     domain.FuelType      `json:"fuel_type"`
	Transmission domain.Transmission  `json:"transmission"`
	Drivetrain   string               `json:"drivetrain"`
	EngineSize   string               `json:"engine_size"`
	Doors        int                  `json:"doors"`
	Color        string               `json:"color"`
	ThumbnailURL string               `json:"thumbnail_url"` // first image only
	ViewCount    int                  `json:"view_count"`
	SellerType   domain.SellerType    `json:"seller_type"`
	IsVerified   bool                 `json:"is_verified"` // seller verified badge

	// Buyer review state — lets a seller see why their listing is not live yet
	VerificationStatus domain.VerificationStatus `json:"verification_status"`
	RejectionReason    string                    `json:"rejection_reason,omitempty"`

	CreatedAt time.Time `json:"created_at"`
}

// ToListingResponse maps a domain.Listing to the full response DTO.
func ToListingResponse(l domain.Listing) ListingResponse {
	images := []string{}
	if l.Images != nil {
		images = l.Images
	}

	res := ListingResponse{
		ID:           l.ID,
		Title:        l.Title,
		Description:  l.Description,
		PriceKES:     l.PriceKES,
		Location:     l.Location,
		BodyType:     l.BodyType,
		Status:       l.Status,
		SellerType:   l.SellerType,
		Make:         l.Make,
		Model:        l.Model,
		Year:         l.Year,
		Mileage:      l.Mileage,
		FuelType:     l.FuelType,
		Transmission: l.Transmission,
		Drivetrain:   l.Drivetrain,
		EngineSize:   l.EngineSize,
		Doors:        l.Doors,
		Color:        l.Color,
		Images:       images,
		CoverImage:   l.CoverImage,
		ViewCount:    l.ViewCount,
		Seller:       ToUserSummary(l.User),
		CreatedAt:    l.CreatedAt,
		UpdatedAt:    l.UpdatedAt,

		VerificationStatus:   l.VerificationStatus,
		VerificationFullName: l.VerificationFullName,
		VerificationIDNumber: l.VerificationIDNumber,
		VerificationELogbook: l.VerificationELogbook,
		VerifiedAt:           l.VerifiedAt,
		RejectionReason:      l.RejectionReason,
	}

	if l.User.DealerProfile != nil {
		s := ToDealerSummary(*l.User.DealerProfile)
		res.DealerProfile = &s
	}

	return res
}

// CoverThumbnail picks the photo that represents a listing on cards and in
// search results: the seller's chosen cover when it is still one of the
// listing's images, otherwise the first image.
func CoverThumbnail(l domain.Listing) string {
	if l.CoverImage != "" {
		for _, img := range l.Images {
			if img == l.CoverImage {
				return l.CoverImage
			}
		}
	}
	if len(l.Images) > 0 {
		return l.Images[0]
	}
	return ""
}

// AdminListingResponse is a ListingResponse plus the seller's email address.
// Reviewing an ownership claim needs a way to trace the account, but the public
// listing endpoints must never expose seller emails, so this shape is produced
// only for admin-only routes.
type AdminListingResponse struct {
	ListingResponse
	SellerEmail string `json:"seller_email"`
}

// ToAdminListingResponse maps a domain.Listing to the admin review DTO.
func ToAdminListingResponse(l domain.Listing) AdminListingResponse {
	return AdminListingResponse{
		ListingResponse: ToListingResponse(l),
		SellerEmail:     l.User.Email,
	}
}

// ToListingCardResponse maps a domain.Listing to the lightweight card DTO.
func ToListingCardResponse(l domain.Listing) ListingCardResponse {
	thumbnail := CoverThumbnail(l)

	return ListingCardResponse{
		ID:           l.ID,
		Title:        l.Title,
		PriceKES:     l.PriceKES,
		Location:     l.Location,
		BodyType:     l.BodyType,
		Status:       l.Status,
		Make:         l.Make,
		Model:        l.Model,
		Year:         l.Year,
		Mileage:      l.Mileage,
		FuelType:     l.FuelType,
		Transmission: l.Transmission,
		Drivetrain:   l.Drivetrain,
		EngineSize:   l.EngineSize,
		Doors:        l.Doors,
		Color:        l.Color,
		ThumbnailURL: thumbnail,
		ViewCount:    l.ViewCount,
		SellerType:   l.SellerType,
		IsVerified:   l.User.IsVerified,

		VerificationStatus: l.VerificationStatus,
		RejectionReason:    l.RejectionReason,

		CreatedAt: l.CreatedAt,
	}
}

// SetCoverImageRequest selects the photo used as the listing's card thumbnail.
type SetCoverImageRequest struct {
	// ImageURL must be one of the listing's own uploaded images. An empty
	// value clears the choice and falls back to the first image.
	ImageURL string `json:"image_url" binding:"omitempty,max=500"`
}
