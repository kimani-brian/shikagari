package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/shikagari/api/internal/dto"
	"github.com/shikagari/api/internal/service"
	"github.com/shikagari/api/pkg/response"
	"github.com/shikagari/api/pkg/validator"
)

// ListingHandler handles vehicle listing endpoints.
type ListingHandler struct {
	listingService *service.ListingService
	uploadService  *service.UploadService
}

// NewListingHandler constructs a ListingHandler with its dependencies.
func NewListingHandler(
	listingService *service.ListingService,
	uploadService *service.UploadService,
) *ListingHandler {
	return &ListingHandler{
		listingService: listingService,
		uploadService:  uploadService,
	}
}

// Search godoc
// @Summary      Search & filter listings (public)
// @Tags         listings
// @Produce      json
// @Param        search       query  string   false  "Keyword search"
// @Param        location     query  string   false  "Kenyan city"
// @Param        make         query  string   false  "Vehicle make"
// @Param        model        query  string   false  "Vehicle model"
// @Param        min_year     query  int      false  "Minimum year"
// @Param        max_year     query  int      false  "Maximum year"
// @Param        min_price    query  number   false  "Minimum price (KES)"
// @Param        max_price    query  number   false  "Maximum price (KES)"
// @Param        fuel_type    query  string   false  "petrol|diesel|hybrid|electric"
// @Param        transmission query  string   false  "automatic|manual"
// @Param        seller_type  query  string   false  "dealer|private"
// @Param        sort_by      query  string   false  "price_asc|price_desc|year_asc|year_desc|newest"
// @Param        page         query  int      false  "Page number (default: 1)"
// @Param        per_page     query  int      false  "Items per page (default: 20, max: 50)"
// @Success      200  {object}  response.APIResponse{data=[]dto.ListingCardResponse}
// @Router       /listings [get]
func (h *ListingHandler) Search(c *gin.Context) {
	var filters dto.ListingFilterRequest
	if err := c.ShouldBindQuery(&filters); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	cards, total, err := h.listingService.Search(filters)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	meta := h.listingService.BuildSearchMeta(filters, total)
	response.Paginated(c, "Listings retrieved", cards, meta)
}

// GetByID godoc
// @Summary      Get listing by ID (public)
// @Tags         listings
// @Produce      json
// @Param        id  path  string  true  "Listing UUID"
// @Success      200 {object}  response.APIResponse{data=dto.ListingResponse}
// @Failure      404 {object}  response.APIResponse
// @Router       /listings/{id} [get]
func (h *ListingHandler) GetByID(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}

	result, svcErr := h.listingService.GetByID(id)
	if svcErr != nil {
		response.NotFound(c, svcErr.Error())
		return
	}

	response.OK(c, "Listing retrieved", result)
}

// Create godoc
// @Summary      Create a listing (approved sellers only)
// @Tags         listings
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      dto.CreateListingRequest  true  "Listing payload"
// @Success      201   {object}  response.APIResponse{data=dto.ListingResponse}
// @Failure      400   {object}  response.APIResponse
// @Failure      403   {object}  response.APIResponse
// @Router       /listings [post]
func (h *ListingHandler) Create(c *gin.Context) {
	claims := mustGetClaims(c)

	var req dto.CreateListingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if errs := validator.Validate(req); errs != nil {
		response.UnprocessableEntity(c, errs)
		return
	}

	result, err := h.listingService.Create(claims.UserID, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Created(c, "Listing created successfully", result)
}

// Update godoc
// @Summary      Update a listing
// @Tags         listings
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string                    true  "Listing UUID"
// @Param        body  body  dto.UpdateListingRequest  true  "Update payload"
// @Success      200   {object}  response.APIResponse{data=dto.ListingResponse}
// @Router       /listings/{id} [patch]
func (h *ListingHandler) Update(c *gin.Context) {
	claims := mustGetClaims(c)

	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}

	var req dto.UpdateListingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	isAdmin := isAdminUser(c)

	result, svcErr := h.listingService.Update(id, claims.UserID, isAdmin, req)
	if svcErr != nil {
		response.BadRequest(c, svcErr.Error())
		return
	}

	response.OK(c, "Listing updated successfully", result)
}

// Delete godoc
// @Summary      Delete a listing
// @Tags         listings
// @Security     BearerAuth
// @Param        id  path  string  true  "Listing UUID"
// @Success      200 {object}  response.APIResponse
// @Router       /listings/{id} [delete]
func (h *ListingHandler) Delete(c *gin.Context) {
	claims := mustGetClaims(c)

	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}

	isAdmin := isAdminUser(c)

	if err := h.listingService.Delete(id, claims.UserID, isAdmin); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, "Listing deleted successfully", nil)
}

// GetMyListings godoc
// @Summary      Get my listings
// @Tags         listings
// @Security     BearerAuth
// @Param        page      query  int  false  "Page"
// @Param        per_page  query  int  false  "Per page"
// @Success      200  {object}  response.APIResponse{data=[]dto.ListingCardResponse}
// @Router       /listings/me [get]
func (h *ListingHandler) GetMyListings(c *gin.Context) {
	claims := mustGetClaims(c)
	page, perPage := getPagination(c)

	cards, total, err := h.listingService.GetMyListings(claims.UserID, page, perPage)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Paginated(c, "Your listings retrieved", cards, buildPaginationMeta(page, perPage, total))
}

// UploadImages godoc
// @Summary      Upload images for a listing
// @Tags         listings
// @Accept       multipart/form-data
// @Security     BearerAuth
// @Param        id      path      string  true   "Listing UUID"
// @Param        images  formData  file    true   "Image files (up to 10 total)"
// @Success      200     {object}  response.APIResponse{data=dto.ListingResponse}
// @Router       /listings/{id}/images [post]
func (h *ListingHandler) UploadImages(c *gin.Context) {
	claims := mustGetClaims(c)

	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		response.BadRequest(c, "failed to parse multipart form")
		return
	}

	files := form.File["images"]
	if len(files) == 0 {
		response.BadRequest(c, "at least one image file is required")
		return
	}
	if len(files) > 10 {
		response.BadRequest(c, "a maximum of 10 images can be uploaded at once")
		return
	}

	// Upload each image and collect URLs
	var imageURLs []string
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			response.BadRequest(c, "failed to open uploaded file")
			return
		}
		defer file.Close()

		url, uploadErr := h.uploadService.UploadImage(file, fileHeader, "listings")
		if uploadErr != nil {
			response.BadRequest(c, uploadErr.Error())
			return
		}
		imageURLs = append(imageURLs, url)
	}

	result, svcErr := h.listingService.AddImages(id, claims.UserID, imageURLs)
	if svcErr != nil {
		response.BadRequest(c, svcErr.Error())
		return
	}

	response.OK(c, "Images uploaded successfully", result)
}

// SetCoverImage godoc
// @Summary      Choose which photo is used as the listing thumbnail
// @Tags         listings
// @Accept       json
// @Security     BearerAuth
// @Param        id    path      string  true   "Listing UUID"
// @Param        body  body      dto.SetCoverImageRequest  true  "Cover image"
// @Success      200   {object}  response.APIResponse{data=dto.ListingResponse}
// @Router       /listings/{id}/cover [patch]
func (h *ListingHandler) SetCoverImage(c *gin.Context) {
	claims := mustGetClaims(c)

	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}

	var req dto.SetCoverImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, svcErr := h.listingService.SetCoverImage(id, claims.UserID, isAdminUser(c), req.ImageURL)
	if svcErr != nil {
		response.BadRequest(c, svcErr.Error())
		return
	}

	response.OK(c, "Cover image updated successfully", result)
}

// UploadELogbook godoc
// @Summary      Upload the NTSA e-logbook for seller verification
// @Tags         listings
// @Accept       multipart/form-data
// @Security     BearerAuth
// @Param        id    path      string  true  "Listing UUID"
// @Param        elogbook  formData  file    true  "NTSA e-logbook image or PDF"
// @Success      200  {object}  response.APIResponse{data=dto.ListingResponse}
// @Router       /listings/{id}/elogbook [post]
func (h *ListingHandler) UploadELogbook(c *gin.Context) {
	claims := mustGetClaims(c)

	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}

	file, header, err := c.Request.FormFile("elogbook")
	if err != nil {
		response.BadRequest(c, "NTSA e-logbook file is required")
		return
	}
	defer file.Close()

	url, uploadErr := h.uploadService.UploadELogbook(file, header)
	if uploadErr != nil {
		response.BadRequest(c, uploadErr.Error())
		return
	}

	if err := h.listingService.AttachELogbook(id, claims.UserID, url); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, "NTSA e-logbook uploaded successfully", gin.H{"verification_elogbook_url": url})
}

// AdminListPendingListings godoc
// @Summary      List buyer listings awaiting verification (admin)
// @Tags         admin
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.APIResponse
// @Router       /admin/listings/pending [get]
func (h *ListingHandler) AdminListPendingListings(c *gin.Context) {
	page, perPage := getPagination(c)

	listings, total, err := h.listingService.ListPendingListings(page, perPage)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Paginated(c, "Pending listings retrieved", listings, buildPaginationMeta(page, perPage, total))
}

// AdminVerifyListing godoc
// @Summary      Approve or reject a pending listing (admin)
// @Tags         admin
// @Accept       json
// @Security     BearerAuth
// @Param        id    path      string  true  "Listing UUID"
// @Param        body  body      dto.AdminReviewListingRequest  true  "Review decision"
// @Success      200  {object}  response.APIResponse{data=dto.ListingResponse}
// @Router       /admin/listings/{id}/verify [patch]
func (h *ListingHandler) AdminVerifyListing(c *gin.Context) {
	claims := mustGetClaims(c)

	id, err := parseUUID(c, "id")
	if err != nil {
		return
	}

	var req dto.AdminReviewListingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, validator.Validate(req))
		return
	}

	result, err := h.listingService.AdminReviewListing(id, claims.UserID, req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, "Listing reviewed", result)
}
