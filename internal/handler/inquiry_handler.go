package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/shikagari/api/internal/dto"
	"github.com/shikagari/api/internal/service"
	"github.com/shikagari/api/pkg/response"
	"github.com/shikagari/api/pkg/validator"
)

// InquiryHandler handles buyer-seller inquiry endpoints.
type InquiryHandler struct {
	inquiryService *service.InquiryService
}

// NewInquiryHandler constructs an InquiryHandler.
func NewInquiryHandler(inquiryService *service.InquiryService) *InquiryHandler {
	return &InquiryHandler{inquiryService: inquiryService}
}

// Send godoc
// @Summary      Send inquiry to seller
// @Description  Authenticated buyers can send a message about an active listing
// @Tags         inquiries
// @Accept       json
// @Security     BearerAuth
// @Param        id    path  string                    true  "Listing UUID"
// @Param        body  body  dto.CreateInquiryRequest  true  "Inquiry payload"
// @Success      201   {object}  response.APIResponse{data=dto.InquiryResponse}
// @Failure      400   {object}  response.APIResponse
// @Router       /listings/{id}/inquiries [post]
func (h *InquiryHandler) Send(c *gin.Context) {
	claims := mustGetClaims(c)

	listingID, err := parseUUID(c, "id")
	if err != nil {
		return
	}

	var req dto.CreateInquiryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if errs := validator.Validate(req); errs != nil {
		response.UnprocessableEntity(c, errs)
		return
	}

	result, svcErr := h.inquiryService.Send(claims.UserID, listingID, req)
	if svcErr != nil {
		response.BadRequest(c, svcErr.Error())
		return
	}

	response.Created(c, "Inquiry sent successfully", result)
}

// GetByID godoc
// @Summary      Get an inquiry by ID
// @Description  Accessible by the buyer, seller, or admin involved in the inquiry
// @Tags         inquiries
// @Security     BearerAuth
// @Param        id  path  string  true  "Inquiry UUID"
// @Success      200 {object}  response.APIResponse{data=dto.InquiryResponse}
// @Router       /inquiries/{id} [get]
func (h *InquiryHandler) GetByID(c *gin.Context) {
	claims := mustGetClaims(c)

	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}

	isAdmin := isAdminUser(c)

	result, svcErr := h.inquiryService.GetByID(id, claims.UserID, isAdmin)
	if svcErr != nil {
		response.Forbidden(c, svcErr.Error())
		return
	}

	response.OK(c, "Inquiry retrieved", result)
}

// GetMyInquiries godoc
// @Summary      Get inquiries I sent (buyer view)
// @Tags         inquiries
// @Security     BearerAuth
// @Param        page      query  int  false  "Page"
// @Param        per_page  query  int  false  "Per page"
// @Success      200  {object}  response.APIResponse{data=[]dto.InquirySummary}
// @Router       /inquiries/sent [get]
func (h *InquiryHandler) GetMyInquiries(c *gin.Context) {
	claims := mustGetClaims(c)
	page, perPage := getPagination(c)

	results, total, err := h.inquiryService.GetMyInquiries(claims.UserID, page, perPage)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Paginated(c, "Sent inquiries retrieved", results, buildPaginationMeta(page, perPage, total))
}

// GetMyInbox godoc
// @Summary      Get inquiries I received (seller inbox)
// @Tags         inquiries
// @Security     BearerAuth
// @Param        page      query  int  false  "Page"
// @Param        per_page  query  int  false  "Per page"
// @Success      200  {object}  response.APIResponse{data=[]dto.InquirySummary}
// @Router       /inquiries/inbox [get]
func (h *InquiryHandler) GetMyInbox(c *gin.Context) {
	claims := mustGetClaims(c)
	page, perPage := getPagination(c)

	results, total, err := h.inquiryService.GetMyInbox(claims.UserID, page, perPage)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Paginated(c, "Inbox retrieved", results, buildPaginationMeta(page, perPage, total))
}

// Reply godoc
// @Summary      Reply to an inquiry (seller only)
// @Tags         inquiries
// @Accept       json
// @Security     BearerAuth
// @Param        id    path  string                   true  "Inquiry UUID"
// @Param        body  body  dto.ReplyInquiryRequest  true  "Reply payload"
// @Success      200   {object}  response.APIResponse{data=dto.InquiryResponse}
// @Router       /inquiries/{id}/reply [patch]
func (h *InquiryHandler) Reply(c *gin.Context) {
	claims := mustGetClaims(c)

	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}

	var req dto.ReplyInquiryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if errs := validator.Validate(req); errs != nil {
		response.UnprocessableEntity(c, errs)
		return
	}

	result, svcErr := h.inquiryService.Reply(id, claims.UserID, req)
	if svcErr != nil {
		response.BadRequest(c, svcErr.Error())
		return
	}

	response.OK(c, "Reply sent successfully", result)
}

// UpdateStatus godoc
// @Summary      Update inquiry status (seller only)
// @Tags         inquiries
// @Accept       json
// @Security     BearerAuth
// @Param        id    path  string                          true  "Inquiry UUID"
// @Param        body  body  dto.UpdateInquiryStatusRequest  true  "Status payload"
// @Success      200   {object}  response.APIResponse{data=dto.InquiryResponse}
// @Router       /inquiries/{id}/status [patch]
func (h *InquiryHandler) UpdateStatus(c *gin.Context) {
	claims := mustGetClaims(c)

	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}

	var req dto.UpdateInquiryStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if errs := validator.Validate(req); errs != nil {
		response.UnprocessableEntity(c, errs)
		return
	}

	result, svcErr := h.inquiryService.UpdateStatus(id, claims.UserID, req)
	if svcErr != nil {
		response.BadRequest(c, svcErr.Error())
		return
	}

	response.OK(c, "Inquiry status updated", result)
}
