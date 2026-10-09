package utils

import (
	"nas-go/api/pkg/i18n"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

const MaxPageSize = 500

func ParsePagination(c *gin.Context, defaultPageSize int) (page int, pageSize int, isValid bool) {
	page, isPageValid := parseQueryInt(c, "page", 1)
	if !isPageValid {
		return 0, 0, false
	}

	pageSize, isPageSizeValid := parseQueryInt(c, "page_size", defaultPageSize)
	if !isPageSizeValid {
		return 0, 0, false
	}

	return max(page, 1), min(max(pageSize, 1), MaxPageSize), true
}

func parseQueryInt(c *gin.Context, name string, defaultValue int) (int, bool) {
	rawValue := c.Query(name)
	if rawValue == "" {
		return defaultValue, true
	}

	parsedValue, err := strconv.Atoi(rawValue)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_PAGINATION")})
		return 0, false
	}

	return parsedValue, true
}
