package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shikagari/api/internal/domain"
)

// ── Request DTOs ─────────────────────────────────────────────────────────────

// CreateInquiryRequest is the payload for POST /listings/:id/inquiries
type CreateInquiryRequest struct {
	Message string `json:"message" binding:"required,min=10,max=1000"`
}

// ReplyInquiryRequest is the payload for PATCH /inquiries/:id/reply
// Used by sellers to respond to a buyer's message.
type ReplyInquiryRequest struct {
	Reply string `json:"reply" binding:"required,min=5,max=1000"`
}

// UpdateInquiryStatusRequest allows a seller to close an inquiry.
// Used for PATCH /inquiries/:id/status
type UpdateInquiryStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=open replied closed"`
}

// ── Response DTOs ─────────────────────────────────────────────────────────────

// InquiryResponse is the full inquiry shape for API responses.
type InquiryResponse struct {
	ID        uuid.UUID            `json:"id"`
	ListingID uuid.UUID            `json:"listing_id"`
	Message   string               `json:"message"`
	Reply     string               `json:"reply,omitempty"`
	Status    domain.InquiryStatus `json:"status"`

	// Lightweight embeds — keep payload small
	Listing ListingCardResponse `json:"listing"`
	Buyer   UserSummary         `json:"buyer"`
	Seller  UserSummary         `json:"seller"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// InquirySummary is used in list views (e.g. seller's inbox).
type InquirySummary struct {
	ID        uuid.UUID            `json:"id"`
	ListingID uuid.UUID            `json:"listing_id"`
	Listing   *ListingCardResponse `json:"listing,omitempty"`
	Message   string               `json:"message"`
	Status    domain.InquiryStatus `json:"status"`
	Buyer     UserSummary          `json:"buyer"`
	CreatedAt time.Time            `json:"created_at"`
}

// ToInquiryResponse maps a domain.Inquiry to the full response DTO.
func ToInquiryResponse(i domain.Inquiry) InquiryResponse {
	return InquiryResponse{
		ID:        i.ID,
		ListingID: i.ListingID,
		Message:   i.Message,
		Reply:     i.Reply,
		Status:    i.Status,
		Listing:   ToListingCardResponse(i.Listing),
		Buyer:     ToUserSummary(i.Buyer),
		Seller:    ToUserSummary(i.Seller),
		CreatedAt: i.CreatedAt,
		UpdatedAt: i.UpdatedAt,
	}
}

// ToInquirySummary maps a domain.Inquiry to the lightweight summary DTO.
func ToInquirySummary(i domain.Inquiry) InquirySummary {
	var listingSummary *ListingCardResponse
	if i.Listing.ID != uuid.Nil {
		summary := ToListingCardResponse(i.Listing)
		listingSummary = &summary
	}

	return InquirySummary{
		ID:        i.ID,
		ListingID: i.ListingID,
		Listing:   listingSummary,
		Message:   i.Message,
		Status:    i.Status,
		Buyer:     ToUserSummary(i.Buyer),
		CreatedAt: i.CreatedAt,
	}
}
