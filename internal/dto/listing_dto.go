package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shikagari/api/internal/domain"
)

// ── Request DTOs ─────────────────────────────────────────────────────────────

// CreateListingRequest is the payload for POST /listings
type CreateListingRequest struct {
	Title        string  `json:"title"        binding:"required,min=5,max=255"`
	Description  string  `json:"description"  binding:"omitempty,max=5000"`
	PriceKES     float64 `json:"price_kes"    binding:"required,gte=0"`
	Location     string  `json:"location"     binding:"required,oneof=Nairobi Mombasa Kisumu Nakuru Eldoret Thika Malindi Nyeri Machakos Kisii Kericho Garissa Meru Kakamega Other"`
	BodyType     string  `json:"body_type"    binding:"required,oneof=SUV Sedan Hatchback Pickup Coupe EV Van Wagon"`
	Make         string  `json:"make"         binding:"required,min=1,max=100"`
	Model        string  `json:"model"        binding:"required,min=1,max=100"`
	Year         int     `json:"year"         binding:"required,gte=1980,lte=2025"`
	Mileage      int     `json:"mileage"      binding:"required,gte=0"`
	FuelType     string  `json:"fuel_type"    binding:"required,oneof=petrol diesel hybrid electric"`
	Transmission string  `json:"transmission" binding:"required,oneof=automatic manual"`
	Color        string  `json:"color"        binding:"omitempty,max=50"`
	// Images are uploaded separately via POST /listings/:id/images
	// and are not part of the initial create payload.
}

// UpdateListingRequest is the payload for PATCH /listings/:id
// All fields are optional — only provided fields are updated.
type UpdateListingRequest struct {
	Title        *string  `json:"title"        binding:"omitempty,min=5,max=255"`
	Description  *string  `json:"description"  binding:"omitempty,max=5000"`
	PriceKES     *float64 `json:"price_kes"    binding:"omitempty,gte=0"`
	Location     *string  `json:"location"     binding:"omitempty,oneof=Nairobi Mombasa Kisumu Nakuru Eldoret Thika Malindi Nyeri Machakos Kisii Kericho Garissa Meru Kakamega Other"`
	BodyType     *string  `json:"body_type"    binding:"omitempty,oneof=SUV Sedan Hatchback Pickup Coupe EV Van Wagon"`
	Make         *string  `json:"make"         binding:"omitempty,min=1,max=100"`
	Model        *string  `json:"model"        binding:"omitempty,min=1,max=100"`
	Year         *int     `json:"year"         binding:"omitempty,gte=1980,lte=2025"`
	Mileage      *int     `json:"mileage"      binding:"omitempty,gte=0"`
	FuelType     *string  `json:"fuel_type"    binding:"omitempty,oneof=petrol diesel hybrid electric"`
	Transmission *string  `json:"transmission" binding:"omitempty,oneof=automatic manual"`
	Color        *string  `json:"color"        binding:"omitempty,max=50"`
	Status       *string  `json:"status"       binding:"omitempty,oneof=active inactive sold"`
}

// ListingFilterRequest maps query parameters for GET /listings
type ListingFilterRequest struct {
	Search       string  `form:"search"`
	Location     string  `form:"location"`
	BodyType     string  `form:"body_type"      binding:"omitempty,oneof=SUV Sedan Hatchback Pickup Coupe EV Van Wagon"`
	Make         string  `form:"make"`
	Model        string  `form:"model"`
	MinYear      int     `form:"min_year"`
	MaxYear      int     `form:"max_year"`
	MinPriceKES  float64 `form:"min_price"`
	MaxPriceKES  float64 `form:"max_price"`
	FuelType     string  `form:"fuel_type"    binding:"omitempty,oneof=petrol diesel hybrid electric"`
	Transmission string  `form:"transmission" binding:"omitempty,oneof=automatic manual"`
	SellerType   string  `form:"seller_type"  binding:"omitempty,oneof=dealer private"`
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
	Color        string               `json:"color"`
	Images       []string             `json:"images"`
	ViewCount    int                  `json:"view_count"`
	Seller       UserSummary          `json:"seller"`

	// Dealer or private seller profile summary (only one will be non-nil)
	DealerProfile        *DealerSummary        `json:"dealer_profile,omitempty"`
	PrivateSellerProfile *PrivateSellerSummary `json:"private_seller_profile,omitempty"`

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
	ThumbnailURL string               `json:"thumbnail_url"` // first image only
	ViewCount    int                  `json:"view_count"`
	SellerType   domain.SellerType    `json:"seller_type"`
	IsVerified   bool                 `json:"is_verified"` // seller verified badge
	CreatedAt    time.Time            `json:"created_at"`
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
		Color:        l.Color,
		Images:       images,
		ViewCount:    l.ViewCount,
		Seller:       ToUserSummary(l.User),
		CreatedAt:    l.CreatedAt,
		UpdatedAt:    l.UpdatedAt,
	}

	if l.User.DealerProfile != nil {
		s := ToDealerSummary(*l.User.DealerProfile)
		res.DealerProfile = &s
	}

	if l.User.PrivateSellerProfile != nil {
		s := ToPrivateSellerSummary(*l.User.PrivateSellerProfile)
		res.PrivateSellerProfile = &s
	}

	return res
}

// ToListingCardResponse maps a domain.Listing to the lightweight card DTO.
func ToListingCardResponse(l domain.Listing) ListingCardResponse {
	thumbnail := ""
	if len(l.Images) > 0 {
		thumbnail = l.Images[0]
	}

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
		ThumbnailURL: thumbnail,
		ViewCount:    l.ViewCount,
		SellerType:   l.SellerType,
		IsVerified:   l.User.IsVerified,
		CreatedAt:    l.CreatedAt,
	}
}
