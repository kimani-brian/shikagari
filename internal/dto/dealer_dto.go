package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shikagari/api/internal/domain"
)

// ── Request DTOs ─────────────────────────────────────────────────────────────

// CreateDealerProfileRequest is the payload for POST /dealers/profile
type CreateDealerProfileRequest struct {
	BusinessName  string `json:"business_name"   binding:"required,min=2,max=200"`
	BusinessRegNo string `json:"business_reg_no" binding:"omitempty,max=100"`
	Location      string `json:"location"        binding:"required,oneof=Nairobi Mombasa Kisumu Nakuru Eldoret Thika Malindi Nyeri Machakos Kisii Kericho Garissa Meru Kakamega Other"`
	Address       string `json:"address"         binding:"omitempty,max=500"`
	Description   string `json:"description"     binding:"omitempty,max=1000"`
	KRAPIN        string `json:"kra_pin"         binding:"omitempty,max=20"`
}

// UpdateDealerProfileRequest is the payload for PATCH /dealers/profile
type UpdateDealerProfileRequest struct {
	BusinessName string `json:"business_name" binding:"omitempty,min=2,max=200"`
	Location     string `json:"location"      binding:"omitempty,oneof=Nairobi Mombasa Kisumu Nakuru Eldoret Thika Malindi Nyeri Machakos Kisii Kericho Garissa Meru Kakamega Other"`
	Address      string `json:"address"       binding:"omitempty,max=500"`
	Description  string `json:"description"   binding:"omitempty,max=1000"`
	KRAPIN       string `json:"kra_pin"       binding:"omitempty,max=20"`
}

// AdminReviewDealerRequest is the payload for PATCH /admin/dealers/:id/review
type AdminReviewDealerRequest struct {
	ApprovalStatus string `json:"approval_status" binding:"required,oneof=approved rejected"`
}

// ── Response DTOs ─────────────────────────────────────────────────────────────

// DealerProfileResponse is the full dealer profile shape for API responses.
type DealerProfileResponse struct {
	ID             uuid.UUID             `json:"id"`
	UserID         uuid.UUID             `json:"user_id"`
	BusinessName   string                `json:"business_name"`
	BusinessRegNo  string                `json:"business_reg_no"`
	Location       string                `json:"location"`
	Address        string                `json:"address"`
	LogoURL        string                `json:"logo_url"`
	Description    string                `json:"description"`
	KRAPIN         string                `json:"kra_pin"`
	ApprovalStatus domain.ApprovalStatus `json:"approval_status"`
	ApprovedAt     *time.Time            `json:"approved_at"`
	User           UserSummary           `json:"user"`
	CreatedAt      time.Time             `json:"created_at"`
}

// DealerSummary is a lightweight dealer shape embedded in listing responses.
type DealerSummary struct {
	ID           uuid.UUID `json:"id"`
	BusinessName string    `json:"business_name"`
	Location     string    `json:"location"`
	LogoURL      string    `json:"logo_url"`
	IsVerified   bool      `json:"is_verified"`
}

// ToDealerProfileResponse maps a domain.DealerProfile to the full response DTO.
func ToDealerProfileResponse(d domain.DealerProfile) DealerProfileResponse {
	return DealerProfileResponse{
		ID:             d.ID,
		UserID:         d.UserID,
		BusinessName:   d.BusinessName,
		BusinessRegNo:  d.BusinessRegNo,
		Location:       d.Location,
		Address:        d.Address,
		LogoURL:        d.LogoURL,
		Description:    d.Description,
		KRAPIN:         d.KRAPIN,
		ApprovalStatus: d.ApprovalStatus,
		ApprovedAt:     d.ApprovedAt,
		User:           ToUserSummary(d.User),
		CreatedAt:      d.CreatedAt,
	}
}

// ToDealerSummary maps a domain.DealerProfile to a lightweight listing-embed DTO.
func ToDealerSummary(d domain.DealerProfile) DealerSummary {
	return DealerSummary{
		ID:           d.ID,
		BusinessName: d.BusinessName,
		Location:     d.Location,
		LogoURL:      d.LogoURL,
		IsVerified:   d.User.IsVerified,
	}
}
