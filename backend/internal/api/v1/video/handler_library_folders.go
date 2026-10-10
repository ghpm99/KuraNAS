package video

import (
	"net/http"
	"strings"

	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

const (
	defaultLibraryFolderPageSize = 50
	maxLibraryFolderPathLength   = 4096
)

func parseLibraryFolderPath(c *gin.Context, queryName string) (string, bool) {
	folderPath := strings.TrimSpace(c.Query(queryName))
	isValid := len(folderPath) <= maxLibraryFolderPathLength && !strings.ContainsRune(folderPath, 0)
	return folderPath, isValid
}

func rejectInvalidLibraryFolder(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_VIDEO_INVALID_REQUEST")})
}

func (h *Handler) ListLibraryFoldersHandler(c *gin.Context) {
	page, pageSize, isPaginationValid := utils.ParsePagination(c, defaultLibraryFolderPageSize)
	if !isPaginationValid {
		return
	}
	parentPath, isPathValid := parseLibraryFolderPath(c, "parent")
	if !isPathValid {
		rejectInvalidLibraryFolder(c)
		return
	}

	folderPage, err := h.service.ListLibraryFolders(LibraryFolderRequest{ParentPath: parentPath, Page: page, PageSize: pageSize})
	if err != nil {
		respondVideoError(c, err)
		return
	}
	c.JSON(http.StatusOK, folderPage)
}

func (h *Handler) ListLibraryFolderVideosHandler(c *gin.Context) {
	page, pageSize, isPaginationValid := utils.ParsePagination(c, defaultLibraryFolderPageSize)
	if !isPaginationValid {
		return
	}
	folderPath, isPathValid := parseLibraryFolderPath(c, "path")
	if !isPathValid || folderPath == "" {
		rejectInvalidLibraryFolder(c)
		return
	}

	videoPage, err := h.service.ListLibraryFolderVideos(LibraryFolderVideosRequest{FolderPath: folderPath, Page: page, PageSize: pageSize})
	if err != nil {
		respondVideoError(c, err)
		return
	}
	c.JSON(http.StatusOK, videoPage)
}
