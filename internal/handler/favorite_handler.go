package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/shikagari/api/internal/service"
	"github.com/shikagari/api/pkg/response"
)

// FavoriteHandler handles wishlist/favorites endpoints.
type FavoriteHandler struct {
	favoriteService *service.FavoriteService
}

// NewFavoriteHandler constructs a FavoriteHandler.
func NewFavoriteHandler(favoriteService *service.FavoriteService) *FavoriteHandler {
	return &FavoriteHandler{favoriteService: favoriteService}
}

// Toggle godoc
// @Summary      Toggle a listing in favorites
// @Description  Adds the listing if not saved; removes it if already saved
// @Tags         favorites
// @Security     BearerAuth
// @Param        id  path  string  true  "Listing UUID"
// @Success      200 {object}  response.APIResponse{data=dto.FavoriteToggleResponse}
// @Router       /favorites/{id}/toggle [post]
func (h *FavoriteHandler) Toggle(c *gin.Context) {
	claims := mustGetClaims(c)

	listingID, err := parseUUID(c, "id")
	if err != nil {
		return
	}

	result, svcErr := h.favoriteService.Toggle(claims.UserID, listingID)
	if svcErr != nil {
		response.BadRequest(c, svcErr.Error())
		return
	}

	response.OK(c, result.Message, result)
}

// GetMyFavorites godoc
// @Summary      Get my saved listings
// @Tags         favorites
// @Security     BearerAuth
// @Param        page      query  int  false  "Page"
// @Param        per_page  query  int  false  "Per page"
// @Success      200  {object}  response.APIResponse{data=[]dto.FavoriteResponse}
// @Router       /favorites [get]
func (h *FavoriteHandler) GetMyFavorites(c *gin.Context) {
	claims := mustGetClaims(c)
	page, perPage := getPagination(c)

	results, total, err := h.favoriteService.GetMyFavorites(claims.UserID, page, perPage)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Paginated(c, "Favorites retrieved", results, buildPaginationMeta(page, perPage, total))
}
