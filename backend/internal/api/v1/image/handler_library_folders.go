package image

import (
	"net/http"
	"strings"

	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

const maxLibraryFolderPathLength = 4096

func parseLibraryFolderRequest(c *gin.Context) (LibraryFolderRequest, error) {
	parentPath := strings.TrimSpace(c.Query("parent"))
	if len(parentPath) > maxLibraryFolderPathLength || strings.ContainsRune(parentPath, 0) {
		return LibraryFolderRequest{}, errInvalidLibraryFolder
	}
	return LibraryFolderRequest{ParentPath: parentPath}, nil
}

// ListLibraryFoldersHandler serves GET /image/library/folders.
func (h *LibraryHandler) ListLibraryFoldersHandler(c *gin.Context) {
	loggerModel := h.startLog(c, "ListLibraryFolders", "Listing gallery folders")

	page, pageSize, isPaginationValid := utils.ParsePagination(c, defaultLibraryFolderPageSize)
	if !isPaginationValid {
		return
	}

	request, err := parseLibraryFolderRequest(c)
	if err != nil {
		h.rejectInvalidRequest(c, loggerModel, err)
		return
	}
	request.Page = page
	request.PageSize = pageSize

	folderPage, err := h.service.ListLibraryFolders(request)
	if err != nil {
		h.failInternally(c, loggerModel, err)
		return
	}

	h.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, folderPage)
}
