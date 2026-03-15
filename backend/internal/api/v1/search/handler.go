package search

import (
	"errors"
	"nas-go/api/pkg/i18n"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service ServiceInterface
}

func NewHandler(service ServiceInterface) *Handler {
	return &Handler{service: service}
}

func (handler *Handler) SearchGlobalHandler(c *gin.Context) {
	if handler.service == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_CONFIGURATION_LOAD_FAILED")})
		return
	}

	query := strings.TrimSpace(c.Query("q"))
	limit, err := parseLimit(c.Query("limit"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	response, err := handler.service.SearchGlobal(query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_CONFIGURATION_LOAD_FAILED")})
		return
	}

	c.JSON(http.StatusOK, response)
}

func parseLimit(value string) (int, error) {
	if strings.TrimSpace(value) == "" {
		return defaultSearchLimit, nil
	}

	limit, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	if limit < 0 {
		return 0, errors.New("limit must be positive")
	}
	return limit, nil
}
