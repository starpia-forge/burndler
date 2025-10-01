package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ValidationErrorItem represents a single field validation error
type ValidationErrorItem struct {
	Field   string
	Message string
}

// BadRequest sends a 400 Bad Request response with the given message
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, ErrorResponse{
		Error:   "BAD_REQUEST",
		Message: message,
	})
}

// BadRequestWithDetails sends a 400 Bad Request response with details
func BadRequestWithDetails(c *gin.Context, message string, details map[string]string) {
	c.JSON(http.StatusBadRequest, ErrorResponse{
		Error:   "BAD_REQUEST",
		Message: message,
		Details: details,
	})
}

// NotFound sends a 404 Not Found response for the specified resource
func NotFound(c *gin.Context, resource string) {
	c.JSON(http.StatusNotFound, ErrorResponse{
		Error:   "NOT_FOUND",
		Message: resource + " not found",
	})
}

// InternalError sends a 500 Internal Server Error response
func InternalError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error:   "INTERNAL_ERROR",
		Message: message,
	})
}

// ValidationErrors sends a 400 Bad Request response with validation errors
func ValidationErrors(c *gin.Context, errors []ValidationErrorItem) {
	details := make(map[string]string)
	for _, err := range errors {
		details[err.Field] = err.Message
	}

	c.JSON(http.StatusBadRequest, ErrorResponse{
		Error:   "VALIDATION_ERROR",
		Message: "Validation failed",
		Details: details,
	})
}

// Unauthorized sends a 401 Unauthorized response
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, ErrorResponse{
		Error:   "UNAUTHORIZED",
		Message: message,
	})
}

// Forbidden sends a 403 Forbidden response
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, ErrorResponse{
		Error:   "FORBIDDEN",
		Message: message,
	})
}

// Conflict sends a 409 Conflict response
func Conflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, ErrorResponse{
		Error:   "CONFLICT",
		Message: message,
	})
}
