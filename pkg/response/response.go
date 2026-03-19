package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIResponse is the standard envelope for all ShikaGari API responses.
// Keeping a consistent shape makes client-side parsing predictable.
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"` // used for paginated responses
}

// Meta carries pagination information for list endpoints.
type Meta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

// ── Success Responses ─────────────────────────────────────────────────────────

func OK(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Created(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Paginated(c *gin.Context, message string, data interface{}, meta *Meta) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

func NoContent(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}

// ── Error Responses ───────────────────────────────────────────────────────────

func BadRequest(c *gin.Context, err interface{}) {
	c.JSON(http.StatusBadRequest, APIResponse{
		Success: false,
		Message: "Bad request",
		Error:   err,
	})
}

func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, APIResponse{
		Success: false,
		Message: message,
	})
}

func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, APIResponse{
		Success: false,
		Message: message,
	})
}

func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, APIResponse{
		Success: false,
		Message: message,
	})
}

func Conflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, APIResponse{
		Success: false,
		Message: message,
	})
}

func InternalServerError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, APIResponse{
		Success: false,
		Message: message,
	})
}

func UnprocessableEntity(c *gin.Context, err interface{}) {
	c.JSON(http.StatusUnprocessableEntity, APIResponse{
		Success: false,
		Message: "Validation failed",
		Error:   err,
	})
}
