package documenttext

import (
	"errors"
	"net/http"
	"strconv"
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

	filter, folderID, filterErrorKey := parseDocumentFilter(c)
	if filterErrorKey != "" {
		respondDocumentBadRequest(c, filterErrorKey)
		return
	}
	page, pageSize, isPaginationValid := utils.ParsePagination(c, defaultSearchPageSize)
	if !isPaginationValid {
		return
	}

	response, err := handler.service.SearchDocuments(DocumentSearchParams{
		Query:    query,
		Filter:   filter,
		FolderID: folderID,
		Page:     page,
		PageSize: min(pageSize, maxSearchPageSize),
	})
	if errors.Is(err, ErrFolderNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_FILE_NOT_FOUND")})
		return
	}
	if err != nil {
		applog.ErrorWithStack("documenttext: search failed", err, "ip", c.ClientIP())
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_DOCUMENT_SEARCH_FAILED")})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (handler *Handler) GetIndexStateHandler(c *gin.Context) {
	fileID, isValidID := parseFileID(c)
	if !isValidID {
		return
	}

	indexState, err := handler.service.GetIndexState(fileID)
	if errors.Is(err, ErrDocumentNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_DOCUMENT_NOT_FOUND")})
		return
	}
	if err != nil {
		applog.ErrorWithStack("documenttext: index state failed", err, "file_id", fileID, "ip", c.ClientIP())
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_DOCUMENT_INDEX_STATE_FAILED")})
		return
	}

	c.JSON(http.StatusOK, indexState)
}

func (handler *Handler) ReindexDocumentHandler(c *gin.Context) {
	fileID, isValidID := parseFileID(c)
	if !isValidID {
		return
	}

	jobID, err := handler.service.ReindexDocument(fileID)
	if errors.Is(err, ErrDocumentNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_DOCUMENT_NOT_FOUND")})
		return
	}
	if err != nil {
		applog.ErrorWithStack("documenttext: reindex failed", err, "file_id", fileID, "ip", c.ClientIP())
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_DOCUMENT_REINDEX_FAILED")})
		return
	}

	c.JSON(http.StatusAccepted, ReindexJobDto{JobID: jobID})
}

func parseFileID(c *gin.Context) (int, bool) {
	fileID, err := strconv.Atoi(c.Param("file_id"))
	if err != nil || fileID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return 0, false
	}
	return fileID, true
}
