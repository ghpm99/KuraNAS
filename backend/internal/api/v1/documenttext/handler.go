package documenttext

import (
	"net/http"
	"strings"
	"unicode/utf8"

	"nas-go/api/pkg/applog"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

const (
	defaultSearchPageSize = 20
	maxSearchPageSize     = 50
	minQueryLength        = 2
	maxQueryLength        = 200
)

type Handler struct {
	service ServiceInterface
}

func NewHandler(service ServiceInterface) *Handler {
	return &Handler{service: service}
}

func (handler *Handler) SearchDocumentsHandler(c *gin.Context) {
	if handler.service == nil {
		applog.Error("documenttext: service unavailable", "ip", c.ClientIP())
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_DOCUMENT_SEARCH_FAILED")})
		return
	}

	query := strings.TrimSpace(c.Query("q"))
	queryLength := utf8.RuneCountInString(query)
	if queryLength < minQueryLength || queryLength > maxQueryLength {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	page, pageSize, isPaginationValid := utils.ParsePagination(c, defaultSearchPageSize)
	if !isPaginationValid {
		return
	}
	pageSize = min(pageSize, maxSearchPageSize)

	response, err := handler.service.SearchDocuments(query, page, pageSize)
	if err != nil {
		applog.ErrorWithStack("documenttext: search failed", err, "ip", c.ClientIP())
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_DOCUMENT_SEARCH_FAILED")})
		return
	}

	c.JSON(http.StatusOK, response)
}
