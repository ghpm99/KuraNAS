package documenttext

import (
	"errors"
	"net/http"

	"nas-go/api/pkg/applog"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

const defaultDocumentPageSize = 50

func respondDocumentBadRequest(c *gin.Context, messageKey string) {
	c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage(messageKey)})
}

func (handler *Handler) respondDocumentListFailure(c *gin.Context, err error) {
	if errors.Is(err, ErrFolderNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_FILE_NOT_FOUND")})
		return
	}
	applog.ErrorWithStack("documenttext: document list failed", err, "ip", c.ClientIP())
	c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_DOCUMENT_LIST_FAILED")})
}

func (handler *Handler) ListDocumentsHandler(c *gin.Context) {
	filter, folderID, filterErrorKey := parseDocumentFilter(c)
	if filterErrorKey != "" {
		respondDocumentBadRequest(c, filterErrorKey)
		return
	}
	sort, isSortValid := parseDocumentSort(c.Query("sort"))
	if !isSortValid {
		respondDocumentBadRequest(c, "ERROR_DOCUMENT_INVALID_SORT")
		return
	}
	order, isOrderValid := parseDocumentOrder(c.Query("order"), sort)
	if !isOrderValid {
		respondDocumentBadRequest(c, "ERROR_DOCUMENT_INVALID_ORDER")
		return
	}
	page, pageSize, isPaginationValid := utils.ParsePagination(c, defaultDocumentPageSize)
	if !isPaginationValid {
		return
	}

	response, err := handler.service.ListDocuments(DocumentListParams{
		Filter:   filter,
		FolderID: folderID,
		Sort:     sort,
		Order:    order,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		handler.respondDocumentListFailure(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (handler *Handler) CountDocumentsHandler(c *gin.Context) {
	filter, folderID, filterErrorKey := parseDocumentFilter(c)
	if filterErrorKey != "" {
		respondDocumentBadRequest(c, filterErrorKey)
		return
	}

	count, err := handler.service.CountDocuments(filter, folderID)
	if err != nil {
		handler.respondDocumentListFailure(c, err)
		return
	}
	c.JSON(http.StatusOK, count)
}
