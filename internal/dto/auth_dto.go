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

// SessionMetadata captures contextual information about a login session.
type SessionMetadata struct {
	UserAgent string
	IPAddress string
	Device    string
	Browser   string
	Location  string
}

// PasswordResetRequest is the payload for POST /auth/password/reset-request
type PasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// PasswordResetConfirmRequest is the payload for POST /auth/password/reset
type PasswordResetConfirmRequest struct {
	Token       string `json:"token"        binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=72"`
}

// ── Response DTOs ─────────────────────────────────────────────────────────────

// AuthResponse is returned on successful register or login.
type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// SecurityEventResponse represents an entry in the security activity feed.
type SecurityEventResponse struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	Timestamp string `json:"timestamp"`
	Severity  string `json:"severity"`
	Details   string `json:"details"`
}

// SessionResponse represents an active session/device for a user.
type SessionResponse struct {
	ID         string `json:"id"`
	Device     string `json:"device"`
	Browser    string `json:"browser"`
	Location   string `json:"location"`
	LastActive string `json:"last_active"`
	Current    bool   `json:"current"`
}
