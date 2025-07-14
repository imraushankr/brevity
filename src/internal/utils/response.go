package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIError returns a standardized error response
func APIError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{
		"success": false,
		"error":   message,
		"data":    nil,
	})
}

// APISuccess returns a standardized success response
func APISuccess(c *gin.Context, statusCode int, data interface{}) {
	response := gin.H{
		"success": true,
		"error":   nil,
	}

	if data != nil {
		response["data"] = data
	}

	c.JSON(statusCode, response)
}

// APIResponse returns a customized API response
func APIResponse(c *gin.Context, statusCode int, success bool, data interface{}, err interface{}) {
	response := gin.H{
		"success": success,
	}

	if data != nil {
		response["data"] = data
	}

	if err != nil {
		response["error"] = err
	} else {
		response["error"] = nil
	}

	c.JSON(statusCode, response)
}

// PaginatedResponse returns a standardized paginated response
func PaginatedResponse(c *gin.Context, statusCode int, data interface{}, pagination interface{}) {
	c.JSON(statusCode, gin.H{
		"success":    true,
		"data":       data,
		"pagination": pagination,
		"error":      nil,
	})
}

// ValidationError returns a standardized validation error response
func ValidationError(c *gin.Context, errors map[string]string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"success": false,
		"error":   "Validation failed",
		"errors":  errors,
	})
}