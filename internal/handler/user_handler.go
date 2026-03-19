package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/shikagari/api/internal/dto"
	"github.com/shikagari/api/internal/middleware"
	"github.com/shikagari/api/internal/service"
	"github.com/shikagari/api/pkg/response"
	"github.com/shikagari/api/pkg/validator"
)

// UserHandler handles user profile endpoints.
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler constructs a UserHandler with its dependencies.
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetMyProfile godoc
// @Summary      Get my profile
// @Tags         users
// @Security     BearerAuth
// @Success      200  {object}  response.APIResponse{data=dto.UserResponse}
// @Router       /users/me [get]
func (h *UserHandler) GetMyProfile(c *gin.Context) {
	claims := mustGetClaims(c)

	result, err := h.userService.GetProfile(claims.UserID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.OK(c, "Profile retrieved", result)
}

// UpdateMyProfile godoc
// @Summary      Update my profile
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.UpdateProfileRequest  true  "Update payload"
// @Success      200   {object}  response.APIResponse{data=dto.UserResponse}
// @Router       /users/me [patch]
func (h *UserHandler) UpdateMyProfile(c *gin.Context) {
	claims := mustGetClaims(c)

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.userService.UpdateProfile(claims.UserID, req)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.OK(c, "Profile updated successfully", result)
}

// ChangePassword godoc
// @Summary      Change password
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  dto.ChangePasswordRequest  true  "Password payload"
// @Success      200   {object}  response.APIResponse
// @Router       /users/me/password [patch]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	claims := mustGetClaims(c)

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if errs := validator.Validate(req); errs != nil {
		response.UnprocessableEntity(c, errs)
		return
	}

	if err := h.userService.ChangePassword(claims.UserID, req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, "Password changed successfully", nil)
}

// ── Admin Handlers ────────────────────────────────────────────────────────────

// AdminListUsers godoc
// @Summary      List all users (admin)
// @Tags         admin
// @Security     BearerAuth
// @Param        page      query  int  false  "Page number"
// @Param        per_page  query  int  false  "Items per page"
// @Success      200  {object}  response.APIResponse{data=[]dto.UserResponse}
// @Router       /admin/users [get]
func (h *UserHandler) AdminListUsers(c *gin.Context) {
	page, perPage := getPagination(c)

	users, total, err := h.userService.AdminListUsers(page, perPage)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Paginated(c, "Users retrieved", users, buildPaginationMeta(page, perPage, total))
}

// AdminGetUser godoc
// @Summary      Get a user by ID (admin)
// @Tags         admin
// @Security     BearerAuth
// @Param        id   path  string  true  "User UUID"
// @Success      200  {object}  response.APIResponse{data=dto.UserResponse}
// @Router       /admin/users/{id} [get]
func (h *UserHandler) AdminGetUser(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}

	result, err := h.userService.AdminGetUser(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.OK(c, "User retrieved", result)
}

// AdminUpdateUser godoc
// @Summary      Update a user (admin)
// @Tags         admin
// @Accept       json
// @Security     BearerAuth
// @Param        id    path  string                    true  "User UUID"
// @Param        body  body  dto.AdminUpdateUserRequest true  "Update payload"
// @Success      200   {object}  response.APIResponse{data=dto.UserResponse}
// @Router       /admin/users/{id} [patch]
func (h *UserHandler) AdminUpdateUser(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}

	var req dto.AdminUpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, svcErr := h.userService.AdminUpdateUser(id, req)
	if svcErr != nil {
		response.BadRequest(c, svcErr.Error())
		return
	}

	response.OK(c, "User updated successfully", result)
}

// AdminDeleteUser godoc
// @Summary      Delete a user (admin)
// @Tags         admin
// @Security     BearerAuth
// @Param        id  path  string  true  "User UUID"
// @Success      200 {object}  response.APIResponse
// @Router       /admin/users/{id} [delete]
func (h *UserHandler) AdminDeleteUser(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}

	if err := h.userService.AdminDeleteUser(id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, "User deleted successfully", nil)
}

// mustGetClaims is a shared helper that extracts JWT claims from context.
// It panics only if the middleware was incorrectly configured (programmer error).
func mustGetClaims(c *gin.Context) *middleware.Claims {
	return middleware.GetCurrentUser(c)
}
