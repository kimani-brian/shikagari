package handler

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shikagari/api/internal/dto"
	"github.com/shikagari/api/pkg/response"
	"github.com/shikagari/api/pkg/validator"
)

// AuthService declares the methods required by the auth handler.
type AuthService interface {
	Register(req dto.RegisterRequest, meta dto.SessionMetadata) (*dto.AuthResponse, error)
	Login(req dto.LoginRequest, meta dto.SessionMetadata) (*dto.AuthResponse, error)
	RequestPasswordReset(req dto.PasswordResetRequest) error
	ResetPassword(req dto.PasswordResetConfirmRequest) error
	ListSecurityEvents(userID uuid.UUID) ([]dto.SecurityEventResponse, error)
	ListSessions(userID uuid.UUID, currentSessionID uuid.UUID) ([]dto.SessionResponse, error)
	RevokeSession(userID uuid.UUID, sessionID uuid.UUID) error
	Logout(userID uuid.UUID, sessionID uuid.UUID) error
}

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	authService AuthService
}

// NewAuthHandler constructs an AuthHandler with its dependencies.
func NewAuthHandler(authService AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register godoc
// @Summary      Register a new user
// @Description  Creates a new buyer or seller account
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.RegisterRequest  true  "Registration payload"
// @Success      201   {object}  response.APIResponse{data=dto.AuthResponse}
// @Failure      400   {object}  response.APIResponse
// @Failure      409   {object}  response.APIResponse
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		// Prefer the per-field map, but fall back to the raw binding error so
		// callers always learn *why* the payload was rejected (e.g. an
		// unsupported role).
		if errs := validator.Validate(req); errs != nil {
			response.BadRequest(c, errs)
			return
		}
		response.BadRequest(c, err.Error())
		return
	}

	if errs := validator.Validate(req); errs != nil {
		response.UnprocessableEntity(c, errs)
		return
	}

	meta := sessionMetadataFromContext(c)
	result, err := h.authService.Register(req, meta)
	if err != nil {
		// Conflict on duplicate email
		if err.Error() == "an account with this email address already exists" {
			response.Conflict(c, err.Error())
			return
		}
		response.InternalServerError(c, err.Error())
		return
	}

	response.Created(c, "Account created successfully. Welcome to ShikaGari!", result)
}

// Login godoc
// @Summary      Login
// @Description  Authenticates a user and returns a JWT
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.LoginRequest  true  "Login payload"
// @Success      200   {object}  response.APIResponse{data=dto.AuthResponse}
// @Failure      400   {object}  response.APIResponse
// @Failure      401   {object}  response.APIResponse
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, validator.Validate(req))
		return
	}

	if errs := validator.Validate(req); errs != nil {
		response.UnprocessableEntity(c, errs)
		return
	}

	meta := sessionMetadataFromContext(c)
	result, err := h.authService.Login(req, meta)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.OK(c, "Login successful", result)
}

// RequestPasswordReset godoc
// @Summary      Request password reset
// @Description  Generates a password reset link for the supplied email
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  dto.PasswordResetRequest  true  "Reset request payload"
// @Success      200   {object}  response.APIResponse
// @Router       /auth/password/reset-request [post]
func (h *AuthHandler) RequestPasswordReset(c *gin.Context) {
	var req dto.PasswordResetRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, validator.Validate(req))
		return
	}

	if errs := validator.Validate(req); errs != nil {
		response.UnprocessableEntity(c, errs)
		return
	}

	if err := h.authService.RequestPasswordReset(req); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.OK(c, "If the email exists, a reset link has been sent", nil)
}

// ResetPassword godoc
// @Summary      Reset password
// @Description  Confirms a reset token and sets a new password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  dto.PasswordResetConfirmRequest  true  "Reset confirmation payload"
// @Success      200   {object}  response.APIResponse
// @Router       /auth/password/reset [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.PasswordResetConfirmRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, validator.Validate(req))
		return
	}

	if errs := validator.Validate(req); errs != nil {
		response.UnprocessableEntity(c, errs)
		return
	}

	if err := h.authService.ResetPassword(req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, "Password reset successfully", nil)
}

// Me godoc
// @Summary      Get current user
// @Description  Returns the profile of the authenticated user decoded from JWT
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.APIResponse{data=dto.UserResponse}
// @Failure      401  {object}  response.APIResponse
// @Router       /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	// Claims are injected by the Authenticate middleware
	claims := mustGetClaims(c)
	response.OK(c, "Authenticated user retrieved", gin.H{
		"id":          claims.UserID,
		"email":       claims.Email,
		"role":        claims.Role,
		"is_verified": claims.IsVerified,
	})
}

// ListSecurityEvents returns recent account security events.
func (h *AuthHandler) ListSecurityEvents(c *gin.Context) {
	claims := mustGetClaims(c)
	events, err := h.authService.ListSecurityEvents(claims.UserID)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.OK(c, "Security events retrieved", events)
}

// ListSessions returns active sessions/devices.
func (h *AuthHandler) ListSessions(c *gin.Context) {
	claims := mustGetClaims(c)
	sessions, err := h.authService.ListSessions(claims.UserID, claims.SessionID)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.OK(c, "Sessions retrieved", sessions)
}

// RevokeSession terminates a session by ID.
func (h *AuthHandler) RevokeSession(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		response.BadRequest(c, gin.H{"session_id": "session id is required"})
		return
	}

	parsedID, err := uuid.Parse(sessionID)
	if err != nil {
		response.BadRequest(c, gin.H{"session_id": "invalid session id"})
		return
	}

	claims := mustGetClaims(c)
	if err := h.authService.RevokeSession(claims.UserID, parsedID); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.OK(c, "Session revoked", gin.H{"session_id": sessionID})
}

// Logout invalidates the caller's current session.
func (h *AuthHandler) Logout(c *gin.Context) {
	claims := mustGetClaims(c)
	if err := h.authService.Logout(claims.UserID, claims.SessionID); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.OK(c, "Logged out", nil)
}

func sessionMetadataFromContext(c *gin.Context) dto.SessionMetadata {
	ua := strings.TrimSpace(c.Request.UserAgent())
	return dto.SessionMetadata{
		UserAgent: ua,
		IPAddress: strings.TrimSpace(c.ClientIP()),
		Device:    strings.TrimSpace(c.GetHeader("X-Device-Name")),
		Browser:   strings.TrimSpace(c.GetHeader("X-Browser-Name")),
		Location:  strings.TrimSpace(c.GetHeader("X-Client-Location")),
	}
}
