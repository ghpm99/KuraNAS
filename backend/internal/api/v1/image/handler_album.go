package image

import (
	"net/http"
	"strconv"

	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

const defaultAlbumPageSize = 60

type AlbumHandler struct {
	service    AlbumServiceInterface
	logService logger.LoggerServiceInterface
}

func NewAlbumHandler(service AlbumServiceInterface, logService logger.LoggerServiceInterface) *AlbumHandler {
	return &AlbumHandler{service: service, logService: logService}
}

func (h *AlbumHandler) startLog(c *gin.Context, name string, description string) logger.LoggerModel {
	loggerModel, _ := h.logService.CreateLog(logger.LoggerModel{
		Name:        name,
		Description: description,
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)
	return loggerModel
}

func (h *AlbumHandler) fail(c *gin.Context, loggerModel logger.LoggerModel, err error) {
	h.logService.CompleteWithErrorLog(loggerModel, err)
	response := albumErrorResponseFor(err)
	c.JSON(response.statusCode, gin.H{"error": i18n.GetMessage(response.messageKey)})
}

func (h *AlbumHandler) succeed(c *gin.Context, loggerModel logger.LoggerModel, statusCode int, body any) {
	h.logService.CompleteWithSuccessLog(loggerModel)
	if body == nil {
		c.Status(statusCode)
		return
	}
	c.JSON(statusCode, body)
}

func parseAlbumID(c *gin.Context) (int, error) {
	albumID, err := strconv.Atoi(c.Param("id"))
	if err != nil || albumID < 1 {
		return 0, ErrAlbumInvalidID
	}
	return albumID, nil
}

// ListAlbumsHandler serves GET /image/albums.
func (h *AlbumHandler) ListAlbumsHandler(c *gin.Context) {
	loggerModel := h.startLog(c, "ListImageAlbums", "Listing image albums")

	page, pageSize, isPaginationValid := utils.ParsePagination(c, defaultAlbumPageSize)
	if !isPaginationValid {
		return
	}

	albums, err := h.service.ListAlbums(AlbumListRequest{Page: page, PageSize: pageSize})
	if err != nil {
		h.fail(c, loggerModel, err)
		return
	}
	h.succeed(c, loggerModel, http.StatusOK, albums)
}

// GetAlbumHandler serves GET /image/albums/:id.
func (h *AlbumHandler) GetAlbumHandler(c *gin.Context) {
	loggerModel := h.startLog(c, "GetImageAlbum", "Reading an image album")

	albumID, err := parseAlbumID(c)
	if err != nil {
		h.fail(c, loggerModel, err)
		return
	}
	album, err := h.service.GetAlbum(albumID)
	if err != nil {
		h.fail(c, loggerModel, err)
		return
	}
	h.succeed(c, loggerModel, http.StatusOK, album)
}

// CreateAlbumHandler serves POST /image/albums.
func (h *AlbumHandler) CreateAlbumHandler(c *gin.Context) {
	loggerModel := h.startLog(c, "CreateImageAlbum", "Creating an image album")

	var body AlbumNameBody
	if err := c.ShouldBindJSON(&body); err != nil {
		h.fail(c, loggerModel, ErrAlbumNameRequired)
		return
	}

	album, err := h.service.CreateAlbum(body.Name)
	if err != nil {
		h.fail(c, loggerModel, err)
		return
	}
	h.succeed(c, loggerModel, http.StatusCreated, album)
}

// UpdateAlbumHandler serves PUT /image/albums/:id.
func (h *AlbumHandler) UpdateAlbumHandler(c *gin.Context) {
	loggerModel := h.startLog(c, "UpdateImageAlbum", "Updating an image album")

	albumID, err := parseAlbumID(c)
	if err != nil {
		h.fail(c, loggerModel, err)
		return
	}
	var body AlbumUpdateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		h.fail(c, loggerModel, ErrAlbumNothingToUpdate)
		return
	}

	album, err := h.service.UpdateAlbum(AlbumUpdate{AlbumID: albumID, Name: body.Name, CoverFileID: body.CoverFileID})
	if err != nil {
		h.fail(c, loggerModel, err)
		return
	}
	h.succeed(c, loggerModel, http.StatusOK, album)
}

// DeleteAlbumHandler serves DELETE /image/albums/:id and never touches the photos.
func (h *AlbumHandler) DeleteAlbumHandler(c *gin.Context) {
	loggerModel := h.startLog(c, "DeleteImageAlbum", "Deleting an image album")

	albumID, err := parseAlbumID(c)
	if err != nil {
		h.fail(c, loggerModel, err)
		return
	}
	if err := h.service.DeleteAlbum(albumID); err != nil {
		h.fail(c, loggerModel, err)
		return
	}
	h.succeed(c, loggerModel, http.StatusNoContent, nil)
}

// AddAlbumItemsHandler serves POST /image/albums/:id/items.
func (h *AlbumHandler) AddAlbumItemsHandler(c *gin.Context) {
	h.changeAlbumItems(c, "AddImageAlbumItems", "Adding photos to an image album", h.service.AddAlbumItems)
}

// RemoveAlbumItemsHandler serves DELETE /image/albums/:id/items.
func (h *AlbumHandler) RemoveAlbumItemsHandler(c *gin.Context) {
	h.changeAlbumItems(c, "RemoveImageAlbumItems", "Removing photos from an image album", h.service.RemoveAlbumItems)
}

func (h *AlbumHandler) changeAlbumItems(c *gin.Context, name string, description string, change func(int, []int) (AlbumItemsChangeDto, error)) {
	loggerModel := h.startLog(c, name, description)

	albumID, err := parseAlbumID(c)
	if err != nil {
		h.fail(c, loggerModel, err)
		return
	}
	var body AlbumFileIDsBody
	if err := c.ShouldBindJSON(&body); err != nil {
		h.fail(c, loggerModel, ErrAlbumInvalidFileIDs)
		return
	}

	itemsChange, err := change(albumID, body.FileIDs)
	if err != nil {
		h.fail(c, loggerModel, err)
		return
	}
	h.succeed(c, loggerModel, http.StatusOK, itemsChange)
}

// ListAlbumItemsHandler serves GET /image/albums/:id/items with the gallery listing parameters.
func (h *AlbumHandler) ListAlbumItemsHandler(c *gin.Context) {
	loggerModel := h.startLog(c, "ListImageAlbumItems", "Listing photos of an image album")

	albumID, err := parseAlbumID(c)
	if err != nil {
		h.fail(c, loggerModel, err)
		return
	}
	page, pageSize, isPaginationValid := utils.ParsePagination(c, defaultLibraryPageSize)
	if !isPaginationValid {
		return
	}
	listing, err := parseLibraryListRequest(c)
	if err != nil {
		h.fail(c, loggerModel, err)
		return
	}
	listing.Page = page
	listing.PageSize = pageSize

	albumPage, err := h.service.ListAlbumItems(AlbumItemsRequest{AlbumID: albumID, Listing: listing})
	if err != nil {
		h.fail(c, loggerModel, err)
		return
	}
	h.succeed(c, loggerModel, http.StatusOK, albumPage)
}
