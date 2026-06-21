package helper

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIResponse is the standard response format for all API endpoints.
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

// SuccessResponse sends a successful JSON response.
func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// ErrorResponse sends an error JSON response.
func ErrorResponse(c *gin.Context, statusCode int, message string, errors interface{}) {
	c.JSON(statusCode, APIResponse{
		Success: false,
		Message: message,
		Errors:  errors,
	})
}

// OK sends a 200 success response.
func OK(c *gin.Context, message string, data interface{}) {
	SuccessResponse(c, http.StatusOK, message, data)
}

// Created sends a 201 success response.
func Created(c *gin.Context, message string, data interface{}) {
	SuccessResponse(c, http.StatusCreated, message, data)
}

// BadRequest sends a 400 error response.
func BadRequest(c *gin.Context, message string, errors interface{}) {
	ErrorResponse(c, http.StatusBadRequest, message, errors)
}

// Unauthorized sends a 401 error response.
func Unauthorized(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusUnauthorized, message, nil)
}

// Forbidden sends a 403 error response.
func Forbidden(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusForbidden, message, nil)
}

// NotFound sends a 404 error response.
func NotFound(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusNotFound, message, nil)
}

// Conflict sends a 409 error response.
func Conflict(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusConflict, message, nil)
}

// InternalError sends a 500 error response.
func InternalError(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusInternalServerError, message, nil)
}

// TooManyRequests sends a 429 error response.
func TooManyRequests(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusTooManyRequests, message, nil)
}
