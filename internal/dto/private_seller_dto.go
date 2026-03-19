package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shikagari/api/internal/domain"
)

// ── Request DTOs ─────────────────────────────────────────────────────────────

// CreatePrivateSellerRequest is the payload for POST /sellers/profile
type CreatePrivateSellerRequest struct {
	NationalIDNo string `json:"national_id_no" binding:"required,min=6,max=20"`
	Location     string `json:"location"       binding:"required,oneof=Nairobi Mombasa Kisumu Nakuru Eldoret Thika Malindi Nyeri Machakos Kisii Kericho Garissa Meru Kakamega Other"`
	Bio          string `json:"bio"            binding:"omitempty,max=500"`
}

// UpdatePrivateSellerRequest is the payload for PATCH /sellers/profile
type UpdatePrivateSellerRequest struct {
	Location string `json:"location" binding:"omitempty,oneof=Nairobi Mombasa Kisumu Nakuru Eldoret Thika Malindi Nyeri Machakos Kisii Kericho Garissa Meru Kakamega Other"`
	Bio      string `json:"bio"      binding:"omitempty,max=500"`
}

// AdminReviewSellerRequest is the payload for PATCH /admin/sellers/:id/review
type AdminReviewSellerRequest struct {
	ApprovalStatus string `json:"approval_status" binding:"required,oneof=approved rejected"`
}

// ── Response DTOs ─────────────────────────────────────────────────────────────

// PrivateSellerProfileResponse is the full private seller shape for API responses.
type PrivateSellerProfileResponse struct {
	ID              uuid.UUID             `json:"id"`
	UserID          uuid.UUID             `json:"user_id"`
	NationalIDNo    string                `json:"national_id_no"`
	Location        string                `json:"location"`
	ProfilePhotoURL string                `json:"profile_photo_url"`
	Bio             string                `json:"bio"`
	ApprovalStatus  domain.ApprovalStatus `json:"approval_status"`
	ApprovedAt      *time.Time            `json:"approved_at"`
	User            UserSummary           `json:"user"`
	CreatedAt       time.Time             `json:"created_at"`
}

// PrivateSellerSummary is a lightweight shape embedded in listing responses.
type PrivateSellerSummary struct {
	ID              uuid.UUID `json:"id"`
	Location        string    `json:"location"`
	ProfilePhotoURL string    `json:"profile_photo_url"`
	IsVerified      bool      `json:"is_verified"`
}

// ToPrivateSellerProfileResponse maps a domain model to the full response DTO.
func ToPrivateSellerProfileResponse(p domain.PrivateSellerProfile) PrivateSellerProfileResponse {
	return PrivateSellerProfileResponse{
		ID:              p.ID,
		UserID:          p.UserID,
		NationalIDNo:    p.NationalIDNo,
		Location:        p.Location,
		ProfilePhotoURL: p.ProfilePhotoURL,
		Bio:             p.Bio,
		ApprovalStatus:  p.ApprovalStatus,
		ApprovedAt:      p.ApprovedAt,
		User:            ToUserSummary(p.User),
		CreatedAt:       p.CreatedAt,
	}
}

// ToPrivateSellerSummary maps a domain model to a lightweight listing-embed DTO.
func ToPrivateSellerSummary(p domain.PrivateSellerProfile) PrivateSellerSummary {
	return PrivateSellerSummary{
		ID:              p.ID,
		Location:        p.Location,
		ProfilePhotoURL: p.ProfilePhotoURL,
		IsVerified:      p.User.IsVerified,
	}
}
