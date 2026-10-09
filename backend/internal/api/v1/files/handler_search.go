package files

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

func parseSearchParentID(rawParentID string) (int, bool) {
	if rawParentID == "" {
		return 0, true
	}
	parentID, err := strconv.Atoi(rawParentID)
	if err != nil || parentID < 0 {
		return 0, false
	}
	return parentID, true
}

func parseSearchIsRecursive(rawRecursive string) (bool, bool) {
	switch rawRecursive {
	case "", "true":
		return true, true
	case "false":
		return false, true
	}
	return false, false
}

func (handler *Handler) SearchFilesHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "SearchFiles",
		Description: "Searching files by name",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	page, pageSize, isPaginationValid := utils.ParsePagination(c, 15)
	if !isPaginationValid {
		return
	}

	query := strings.TrimSpace(c.Query("q"))
	if utf8.RuneCountInString(query) < MinSearchQueryLength {
		handler.respondSearchBadRequest(c, loggerModel, "ERROR_SEARCH_QUERY_TOO_SHORT")
		return
	}

	parentID, isParentValid := parseSearchParentID(c.Query("parent_id"))
	isRecursive, isRecursiveValid := parseSearchIsRecursive(c.Query("recursive"))
	if !isParentValid || !isRecursiveValid {
		handler.respondSearchBadRequest(c, loggerModel, "ERROR_INVALID_REQUEST")
		return
	}

	filter, filterErrorKey := parseSearchFilter(c)
	if filterErrorKey != "" {
		handler.respondSearchBadRequest(c, loggerModel, filterErrorKey)
		return
	}

	loggerModel.SetExtraData(logger.LogExtraData{
		Data: map[string]any{"parent_id": parentID, "recursive": isRecursive, "page": page, "page_size": pageSize},
	})

	pagination, err := handler.service.SearchFilesByName(FileSearchParams{
		Query:       query,
		ParentID:    parentID,
		IsRecursive: isRecursive,
		Page:        page,
		PageSize:    pageSize,
		Filter:      filter,
	})
	if errors.Is(err, ErrFileNotFound) {
		handler.respondNotFound(c, loggerModel, "ERROR_FILE_NOT_FOUND")
		return
	}
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, ParsePaginationToResponse(pagination))
}

func (handler *Handler) respondSearchBadRequest(c *gin.Context, loggerModel logger.LoggerModel, messageKey string) {
	message := i18n.GetMessage(messageKey)
	handler.Logger.CompleteWithErrorLog(loggerModel, errors.New(message))
	c.JSON(http.StatusBadRequest, gin.H{"error": message})
}
