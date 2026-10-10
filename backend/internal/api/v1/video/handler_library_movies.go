package video

import (
	"net/http"

	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

const defaultLibraryMoviePageSize = 24

func (h *Handler) ListLibraryMoviesHandler(c *gin.Context) {
	page, pageSize, isPaginationValid := utils.ParsePagination(c, defaultLibraryMoviePageSize)
	if !isPaginationValid {
		return
	}
	sort, isSortValid := ParseLibraryMovieSort(c.Query("sort"))
	if !isSortValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_VIDEO_INVALID_REQUEST")})
		return
	}

	moviePage, err := h.service.ListLibraryMovies(LibraryMoviesRequest{Sort: sort, Page: page, PageSize: pageSize})
	if err != nil {
		respondVideoError(c, err)
		return
	}
	c.JSON(http.StatusOK, moviePage)
}
