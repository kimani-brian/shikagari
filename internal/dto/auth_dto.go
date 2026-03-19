package dto

// ── Request DTOs ─────────────────────────────────────────────────────────────

// RegisterRequest is the payload for POST /auth/register
type RegisterRequest struct {
	FullName string `json:"full_name" binding:"required,min=2,max=150"`
	Email    string `json:"email"     binding:"required,email"`
	Phone    string `json:"phone"     binding:"required,min=10,max=20"`
	Password string `json:"password"  binding:"required,min=8,max=72"`
	// Role is optional — defaults to "buyer" on the backend.
	// Only "buyer" and "seller" are accepted at registration.
	// "admin" is assigned manually by a superadmin.
	Role string `json:"role"      binding:"omitempty,oneof=buyer seller"`
}

// LoginRequest is the payload for POST /auth/login
type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// ── Response DTOs ─────────────────────────────────────────────────────────────

// AuthResponse is returned on successful register or login.
type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}
