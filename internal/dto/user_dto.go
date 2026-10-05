package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shikagari/api/internal/domain"
)

// ── Request DTOs ─────────────────────────────────────────────────────────────

// UpdateProfileRequest is the payload for PATCH /users/me
type UpdateProfileRequest struct {
	FullName string `json:"full_name" binding:"omitempty,min=2,max=150"`
	Phone    string `json:"phone"     binding:"omitempty,min=10,max=20"`
	Email    string `json:"email"     binding:"omitempty,email,max=255"`
}

// ChangePasswordRequest is the payload for PATCH /users/me/password
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password"     binding:"required,min=8,max=72"`
}

// AdminUpdateUserRequest allows an admin to toggle roles or verified status.
// Used for PATCH /admin/users/:id
type AdminUpdateUserRequest struct {
	Role       *string `json:"role"        binding:"omitempty,oneof=buyer dealer admin"`
	IsVerified *bool   `json:"is_verified" binding:"omitempty"`
	IsActive   *bool   `json:"is_active"   binding:"omitempty"`
}

// ── Response DTOs ─────────────────────────────────────────────────────────────

// UserResponse is the public-safe user shape returned in API responses.
// PasswordHash is never included.
type UserResponse struct {
	ID         uuid.UUID       `json:"id"`
	FullName   string          `json:"full_name"`
	Email      string          `json:"email"`
	Phone      string          `json:"phone"`
	Role       domain.UserRole `json:"role"`
	IsVerified bool            `json:"is_verified"`
	IsActive   bool            `json:"is_active"`
	CreatedAt  time.Time       `json:"created_at"`
}

// UserSummary is a minimal user shape embedded in listings and inquiries
// to keep response payloads lightweight.
type UserSummary struct {
	ID         uuid.UUID       `json:"id"`
	FullName   string          `json:"full_name"`
	Phone      string          `json:"phone"`
	IsVerified bool            `json:"is_verified"`
	Role       domain.UserRole `json:"role"`
}

// ToUserResponse maps a domain.User to a UserResponse DTO.
func ToUserResponse(u domain.User) UserResponse {
	return UserResponse{
		ID:         u.ID,
		FullName:   u.FullName,
		Email:      u.Email,
		Phone:      u.Phone,
		Role:       u.Role,
		IsVerified: u.IsVerified,
		IsActive:   u.IsActive,
		CreatedAt:  u.CreatedAt,
	}
}

// ToUserSummary maps a domain.User to a lightweight UserSummary DTO.
func ToUserSummary(u domain.User) UserSummary {
	return UserSummary{
		ID:         u.ID,
		FullName:   u.FullName,
		Phone:      u.Phone,
		IsVerified: u.IsVerified,
		Role:       u.Role,
	}
}
