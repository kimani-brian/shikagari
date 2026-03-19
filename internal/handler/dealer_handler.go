package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/shikagari/api/internal/dto"
	"github.com/shikagari/api/internal/service"
	"github.com/shikagari/api/pkg/response"
	"github.com/shikagari/api/pkg/validator"
)

// DealerHandler handles dealer profile endpoints.
type DealerHandler struct {
	dealerService *service.DealerService
	uploadService *service.UploadService
}

// NewDealerHandler constructs a DealerHandler with its dependencies.
func NewDealerHandler(
	dealerService *service.DealerService,
	uploadService *service.UploadService,
) *DealerHandler {
	return &DealerHandler{
		dealerService: dealerService,
		uploadService: uploadService,
	}
}

// CreateProfile godoc
// @Summary      Create dealer profile
// @Tags         dealers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.CreateDealerProfileRequest  true  "Dealer profile payload"
// @Success      201   {object}  response.APIResponse{data=dto.DealerProfileResponse}
// @Router       /dealers/profile [post]
func (h *DealerHandler) CreateProfile(c *gin.Context) {
	claims := mustGetClaims(c)

	var req dto.CreateDealerProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if errs := validator.Validate(req); errs != nil {
		response.UnprocessableEntity(c, errs)
		return
	}

	result, err := h.dealerService.CreateProfile(claims.UserID, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Created(c, "Dealer profile created. Awaiting admin approval.", result)
}

// GetMyProfile godoc
// @Summary      Get my dealer profile
// @Tags         dealers
// @Security     BearerAuth
// @Success      200  {object}  response.APIResponse{data=dto.DealerProfileResponse}
// @Router       /dealers/profile [get]
func (h *DealerHandler) GetMyProfile(c *gin.Context) {
	claims := mustGetClaims(c)

	result, err := h.dealerService.GetMyProfile(claims.UserID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.OK(c, "Dealer profile retrieved", result)
}

// GetProfileByID godoc
// @Summary      Get dealer profile by ID (public)
// @Tags         dealers
// @Param        id  path  string  true  "Dealer profile UUID"
// @Success      200 {object}  response.APIResponse{data=dto.DealerProfileResponse}
// @Router       /dealers/{id} [get]
func (h *DealerHandler) GetProfileByID(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}

	result, svcErr := h.dealerService.GetProfileByID(id)
	if svcErr != nil {
		response.NotFound(c, svcErr.Error())
		return
	}

	response.OK(c, "Dealer profile retrieved", result)
}

// UpdateProfile godoc
// @Summary      Update my dealer profile
// @Tags         dealers
// @Accept       json
// @Security     BearerAuth
// @Param        body  body      dto.UpdateDealerProfileRequest  true  "Update payload"
// @Success      200   {object}  response.APIResponse{data=dto.DealerProfileResponse}
// @Router       /dealers/profile [patch]
func (h *DealerHandler) UpdateProfile(c *gin.Context) {
	claims := mustGetClaims(c)

	var req dto.UpdateDealerProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.dealerService.UpdateProfile(claims.UserID, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, "Dealer profile updated", result)
}

// UploadLogo godoc
// @Summary      Upload dealer logo
// @Tags         dealers
// @Accept       multipart/form-data
// @Security     BearerAuth
// @Param        logo  formData  file  true  "Logo image"
// @Success      200   {object}  response.APIResponse{data=dto.UploadLogoResponse}
// @Router       /dealers/profile/logo [post]
func (h *DealerHandler) UploadLogo(c *gin.Context) {
	claims := mustGetClaims(c)

	file, header, err := c.Request.FormFile("logo")
	if err != nil {
		response.BadRequest(c, "logo file is required")
		return
	}
	defer file.Close()

	logoURL, uploadErr := h.uploadService.UploadImage(file, header, "dealers/logos")
	if uploadErr != nil {
		response.BadRequest(c, uploadErr.Error())
		return
	}

	if err := h.dealerService.UpdateLogo(claims.UserID, logoURL); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.OK(c, "Logo uploaded successfully", dto.UploadLogoResponse{LogoURL: logoURL})
}

// ── Admin Handlers ────────────────────────────────────────────────────────────

// AdminListProfiles godoc
// @Summary      List dealer profiles (admin)
// @Tags         admin
// @Security     BearerAuth
// @Param        status    query  string  false  "Filter by approval status (pending/approved/rejected)"
// @Param        page      query  int     false  "Page"
// @Param        per_page  query  int     false  "Per page"
// @Success      200  {object}  response.APIResponse{data=[]dto.DealerProfileResponse}
// @Router       /admin/dealers [get]
func (h *DealerHandler) AdminListProfiles(c *gin.Context) {
	status := c.DefaultQuery("status", "")
	page, perPage := getPagination(c)

	results, total, err := h.dealerService.AdminListProfiles(status, page, perPage)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Paginated(c, "Dealer profiles retrieved", results, buildPaginationMeta(page, perPage, total))
}

// AdminReviewProfile godoc
// @Summary      Approve or reject a dealer profile (admin)
// @Tags         admin
// @Accept       json
// @Security     BearerAuth
// @Param        id    path  string                      true  "Dealer profile UUID"
// @Param        body  body  dto.AdminReviewDealerRequest true  "Review payload"
// @Success      200   {object}  response.APIResponse{data=dto.DealerProfileResponse}
// @Router       /admin/dealers/{id}/review [patch]
func (h *DealerHandler) AdminReviewProfile(c *gin.Context) {
	claims := mustGetClaims(c)

	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}

	var req dto.AdminReviewDealerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if errs := validator.Validate(req); errs != nil {
		response.UnprocessableEntity(c, errs)
		return
	}

	result, svcErr := h.dealerService.AdminReviewProfile(id, claims.UserID, req)
	if svcErr != nil {
		response.BadRequest(c, svcErr.Error())
		return
	}

	response.OK(c, "Dealer profile reviewed successfully", result)
}
