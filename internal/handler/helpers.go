package handler

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shikagari/api/internal/domain"
	"github.com/shikagari/api/internal/middleware"
	"github.com/shikagari/api/pkg/response"
)

// parseUUID extracts and validates a UUID path parameter.
// Writes a 400 response and returns an error if parsing fails.
func parseUUID(c *gin.Context, param string) (uuid.UUID, error) {
	raw := c.Param(param)
	id, err := uuid.Parse(raw)
	if err != nil {
		response.BadRequest(c, "invalid UUID: "+param)
		c.Abort()
		return uuid.Nil, err
	}
	return id, nil
}

// getPagination extracts page and per_page from query params with safe defaults.
func getPagination(c *gin.Context) (int, int) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	perPage, err := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if err != nil || perPage < 1 {
		perPage = 20
	}
	if perPage > 50 {
		perPage = 50
	}

	return page, perPage
}

// buildPaginationMeta constructs a response.Meta struct for paginated endpoints.
func buildPaginationMeta(page, perPage int, total int64) *response.Meta {
	totalPages := int(math.Ceil(float64(total) / float64(perPage)))
	if totalPages < 1 {
		totalPages = 1
	}

	return &response.Meta{
		Page:       page,
		PerPage:    perPage,
		TotalItems: total,
		TotalPages: totalPages,
	}
}

// isAdminUser checks whether the authenticated user has the admin role.
func isAdminUser(c *gin.Context) bool {
	claims := middleware.GetCurrentUser(c)
	if claims == nil {
		return false
	}
	return claims.Role == domain.RoleAdmin
}
