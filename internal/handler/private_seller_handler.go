package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/shikagari/api/internal/dto"
	"github.com/shikagari/api/internal/service"
	"github.com/shikagari/api/pkg/response"
	"github.com/shikagari/api/pkg/validator"
)

// PrivateSellerHandler handles private seller profile endpoints.
type PrivateSellerHandler struct {
	sellerService *service.PrivateSellerService
	uploadService *service.UploadService
}

// NewPrivateSellerHandler constructs a PrivateSellerHandler.
func NewPrivateSellerHandler(
	sellerService *service.PrivateSellerService,
	uploadService *service.UploadService,
) *PrivateSellerHandler {
	return &PrivateSellerHandler{
		sellerService: sellerService,
		uploadService: uploadService,
	}
}

// CreateProfile godoc
// @Summary      Create private seller profile
// @Tags         private-sellers
// @Accept       json
// @Security     BearerAuth
// @Param        body  body      dto.CreatePrivateSellerRequest  true  "Payload"
// @Success      201   {object}  response.APIResponse{data=dto.PrivateSellerProfileResponse}
// @Router       /sellers/profile [post]
func (h *PrivateSellerHandler) CreateProfile(c *gin.Context) {
	claims := mustGetClaims(c)

	var req dto.CreatePrivateSellerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if errs := validator.Validate(req); errs != nil {
		response.UnprocessableEntity(c, errs)
		return
	}

	result, err := h.sellerService.CreateProfile(claims.UserID, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Created(c, "Private seller profile created. Awaiting admin approval.", result)
}

// GetMyProfile godoc
// @Summary      Get my private seller profile
// @Tags         private-sellers
// @Security     BearerAuth
// @Success      200  {object}  response.APIResponse{data=dto.PrivateSellerProfileResponse}
// @Router       /sellers/profile [get]
func (h *PrivateSellerHandler) GetMyProfile(c *gin.Context) {
	claims := mustGetClaims(c)

	result, err := h.sellerService.GetMyProfile(claims.UserID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.OK(c, "Private seller profile retrieved", result)
}

// GetProfileByID godoc
// @Summary      Get private seller profile by ID (public)
// @Tags         private-sellers
// @Param        id  path  string  true  "Profile UUID"
// @Success      200 {object}  response.APIResponse{data=dto.PrivateSellerProfileResponse}
// @Router       /sellers/{id} [get]
func (h *PrivateSellerHandler) GetProfileByID(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}

	result, svcErr := h.sellerService.GetProfileByID(id)
	if svcErr != nil {
		response.NotFound(c, svcErr.Error())
		return
	}

	response.OK(c, "Private seller profile retrieved", result)
}

// UpdateProfile godoc
// @Summary      Update my private seller profile
// @Tags         private-sellers
// @Accept       json
// @Security     BearerAuth
// @Param        body  body      dto.UpdatePrivateSellerRequest  true  "Update payload"
// @Success      200   {object}  response.APIResponse{data=dto.PrivateSellerProfileResponse}
// @Router       /sellers/profile [patch]
func (h *PrivateSellerHandler) UpdateProfile(c *gin.Context) {
	claims := mustGetClaims(c)

	var req dto.UpdatePrivateSellerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.sellerService.UpdateProfile(claims.UserID, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, "Private seller profile updated", result)
}

// UploadProfilePhoto godoc
// @Summary      Upload profile photo
// @Tags         private-sellers
// @Accept       multipart/form-data
// @Security     BearerAuth
// @Param        photo  formData  file  true  "Profile photo"
// @Success      200    {object}  response.APIResponse{data=dto.UploadProfilePhotoResponse}
// @Router       /sellers/profile/photo [post]
func (h *PrivateSellerHandler) UploadProfilePhoto(c *gin.Context) {
	claims := mustGetClaims(c)

	file, header, err := c.Request.FormFile("photo")
	if err != nil {
		response.BadRequest(c, "photo file is required")
		return
	}
	defer file.Close()

	photoURL, uploadErr := h.uploadService.UploadImage(file, header, "sellers/photos")
	if uploadErr != nil {
		response.BadRequest(c, uploadErr.Error())
		return
	}

	if err := h.sellerService.UpdateProfilePhoto(claims.UserID, photoURL); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.OK(c, "Profile photo uploaded successfully", dto.UploadProfilePhotoResponse{
		ProfilePhotoURL: photoURL,
	})
}

// ── Admin Handlers ────────────────────────────────────────────────────────────

// AdminListProfiles godoc
// @Summary      List private seller profiles (admin)
// @Tags         admin
// @Security     BearerAuth
// @Param        status    query  string  false  "Filter: pending/approved/rejected"
// @Param        page      query  int     false  "Page"
// @Param        per_page  query  int     false  "Per page"
// @Success      200  {object}  response.APIResponse
// @Router       /admin/sellers [get]
func (h *PrivateSellerHandler) AdminListProfiles(c *gin.Context) {
	status := c.DefaultQuery("status", "")
	page, perPage := getPagination(c)

	results, total, err := h.sellerService.AdminListProfiles(status, page, perPage)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Paginated(c, "Private seller profiles retrieved", results, buildPaginationMeta(page, perPage, total))
}

// AdminReviewProfile godoc
// @Summary      Approve or reject a private seller profile (admin)
// @Tags         admin
// @Accept       json
// @Security     BearerAuth
// @Param        id    path  string                       true  "Profile UUID"
// @Param        body  body  dto.AdminReviewSellerRequest  true  "Review payload"
// @Success      200   {object}  response.APIResponse
// @Router       /admin/sellers/{id}/review [patch]
func (h *PrivateSellerHandler) AdminReviewProfile(c *gin.Context) {
	claims := mustGetClaims(c)

	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}

	var req dto.AdminReviewSellerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if errs := validator.Validate(req); errs != nil {
		response.UnprocessableEntity(c, errs)
		return
	}

	result, svcErr := h.sellerService.AdminReviewProfile(id, claims.UserID, req)
	if svcErr != nil {
		response.BadRequest(c, svcErr.Error())
		return
	}

	response.OK(c, "Private seller profile reviewed successfully", result)
}
