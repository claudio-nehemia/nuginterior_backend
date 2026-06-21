package helper

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// ParseIDParam parses a uint ID from a URL parameter.
func ParseIDParam(c *gin.Context, paramName string) (uint, error) {
	idStr := c.Param(paramName)
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
