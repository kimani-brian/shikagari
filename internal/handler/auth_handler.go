package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/shikagari/api/internal/dto"
	"github.com/shikagari/api/internal/service"
	"github.com/shikagari/api/pkg/response"
	"github.com/shikagari/api/pkg/validator"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler constructs an AuthHandler with its dependencies.
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
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
		response.BadRequest(c, validator.Validate(req))
		return
	}

	if errs := validator.Validate(req); errs != nil {
		response.UnprocessableEntity(c, errs)
		return
	}

	result, err := h.authService.Register(req)
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

	result, err := h.authService.Login(req)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.OK(c, "Login successful", result)
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
