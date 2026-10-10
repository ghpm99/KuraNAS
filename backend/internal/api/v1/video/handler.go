package video

import (
	"database/sql"
	"errors"
	"nas-go/api/internal/api/v1/clientidentity"
	files "nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/applog"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

const defaultPlaylistItemsPageSize = 50
const defaultPlaylistSectionPageSize = 24

type Handler struct {
	service           ServiceInterface
	filesService      files.ServiceInterface
	recentFileService files.RecentFileServiceInterface
	logService        logger.LoggerServiceInterface
}

func NewHandler(service ServiceInterface, filesService files.ServiceInterface, recentFileService files.RecentFileServiceInterface, logService logger.LoggerServiceInterface) *Handler {
	return &Handler{service: service, filesService: filesService, recentFileService: recentFileService, logService: logService}
}

func respondVideoError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_VIDEO_NOT_FOUND")})
	case errors.Is(err, ErrPlaybackStateNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_VIDEO_PLAYBACK_NOT_FOUND")})
	case errors.Is(err, ErrVideoNotInPlaylist),
		errors.Is(err, ErrInvalidBehaviorEvent),
		errors.Is(err, ErrPlaylistNameRequired),
		errors.Is(err, ErrPlaylistReorderRequired),
		errors.Is(err, ErrPlaybackNavigation),
		errors.Is(err, ErrPlaylistWithoutItems):
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_VIDEO_INVALID_REQUEST")})
	case errors.Is(err, ErrNoVideosForContext):
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_VIDEO_NOT_FOUND")})
	default:
		applog.ErrorWithStack("video: operation failed", err, "path", c.FullPath(), "ip", c.ClientIP())
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_VIDEO_OPERATION_FAILED")})
	}
}

func (h *Handler) StartPlaybackHandler(c *gin.Context) {
	var req StartPlaybackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	clientID, isClientIDValid := clientidentity.Resolve(c)
	if !isClientIDValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	session, err := h.service.StartPlayback(clientID, req.VideoID, req.PlaylistID)
	if err != nil {
		respondVideoError(c, err)
		return
	}

	c.JSON(http.StatusOK, session)
}

func (h *Handler) GetPlaybackStateHandler(c *gin.Context) {
	clientID, isClientIDValid := clientidentity.Resolve(c)
	if !isClientIDValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	session, err := h.service.GetPlaybackState(clientID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_VIDEO_PLAYBACK_NOT_FOUND")})
		return
	}
	c.JSON(http.StatusOK, session)
}

func (h *Handler) UpdatePlaybackStateHandler(c *gin.Context) {
	var req UpdatePlaybackStateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	clientID, isClientIDValid := clientidentity.Resolve(c)
	if !isClientIDValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	state, err := h.service.UpdatePlaybackState(clientID, req)
	if err != nil {
		respondVideoError(c, err)
		return
	}
	c.JSON(http.StatusOK, state)
}

func (h *Handler) NextVideoHandler(c *gin.Context) {
	clientID, isClientIDValid := clientidentity.Resolve(c)
	if !isClientIDValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	session, err := h.service.NextVideo(clientID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_VIDEO_INVALID_REQUEST")})
		return
	}
	c.JSON(http.StatusOK, session)
}

func (h *Handler) PreviousVideoHandler(c *gin.Context) {
	clientID, isClientIDValid := clientidentity.Resolve(c)
	if !isClientIDValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	session, err := h.service.PreviousVideo(clientID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_VIDEO_INVALID_REQUEST")})
		return
	}
	c.JSON(http.StatusOK, session)
}

func (h *Handler) GetHomeCatalogHandler(c *gin.Context) {
	const maxLimit = 100
	limit := utils.ParseInt(c.DefaultQuery("limit", "24"), c)
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}
	if limit <= 0 || limit > maxLimit {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_VIDEO_INVALID_LIMIT")})
		return
	}

	clientID, isClientIDValid := clientidentity.Resolve(c)
	if !isClientIDValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	catalog, err := h.service.GetHomeCatalog(clientID, limit)
	if err != nil {
		respondVideoError(c, err)
		return
	}
	c.JSON(http.StatusOK, catalog)
}

func (h *Handler) GetContinueWatchingHandler(c *gin.Context) {
	const maxLimit = 100
	limit := utils.ParseInt(c.DefaultQuery("limit", "24"), c)
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}
	if limit <= 0 || limit > maxLimit {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_VIDEO_INVALID_LIMIT")})
		return
	}

	clientID, isClientIDValid := clientidentity.Resolve(c)
	if !isClientIDValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	items, err := h.service.GetContinueWatching(clientID, limit)
	if err != nil {
		respondVideoError(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *Handler) SetVideoWatchedHandler(c *gin.Context) {
	videoID := utils.ParseInt(c.Param("file_id"), c)
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	var req SetVideoWatchedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	clientID, isClientIDValid := clientidentity.Resolve(c)
	if !isClientIDValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	if err := h.service.SetVideoWatched(clientID, videoID, *req.Watched); err != nil {
		respondVideoError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) RebuildPlaylistsHandler(c *gin.Context) {
	if err := h.service.RebuildSmartPlaylists(); err != nil {
		respondVideoError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) GetPlaylistsHandler(c *gin.Context) {
	includeHidden := c.DefaultQuery("include_hidden", "false") == "true"
	playlists, err := h.service.GetPlaylists(includeHidden)
	if err != nil {
		respondVideoError(c, err)
		return
	}
	c.JSON(http.StatusOK, playlists)
}

func (h *Handler) GetPlaylistsBySectionHandler(c *gin.Context) {
	section, isSectionValid := ParsePlaylistSection(c.Param("section"))
	if !isSectionValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_VIDEO_INVALID_REQUEST")})
		return
	}
	page, pageSize, isPaginationValid := utils.ParsePagination(c, defaultPlaylistSectionPageSize)
	if !isPaginationValid {
		return
	}

	playlistsPage, err := h.service.GetPlaylistsBySection(PlaylistSectionRequest{Section: section, Page: page, PageSize: pageSize})
	if err != nil {
		respondVideoError(c, err)
		return
	}
	c.JSON(http.StatusOK, playlistsPage)
}

func (h *Handler) GetPlaylistMembershipsHandler(c *gin.Context) {
	includeHidden := c.DefaultQuery("include_hidden", "false") == "true"
	memberships, err := h.service.GetPlaylistMemberships(includeHidden)
	if err != nil {
		respondVideoError(c, err)
		return
	}
	c.JSON(http.StatusOK, memberships)
}

func (h *Handler) GetPlaylistsByVideoHandler(c *gin.Context) {
	videoID := utils.ParseInt(c.Param("file_id"), c)
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	playlists, err := h.service.GetPlaylistsByVideo(videoID)
	if err != nil {
		respondVideoError(c, err)
		return
	}
	c.JSON(http.StatusOK, playlists)
}

func (h *Handler) GetPlaylistByIDHandler(c *gin.Context) {
	id := utils.ParseInt(c.Param("id"), c)
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}
	clientID, isClientIDValid := clientidentity.Resolve(c)
	if !isClientIDValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	playlist, err := h.service.GetPlaylistByID(clientID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_VIDEO_NOT_FOUND")})
		return
	}
	c.JSON(http.StatusOK, playlist)
}

func (h *Handler) GetPlaylistItemsPageHandler(c *gin.Context) {
	id := utils.ParseInt(c.Param("id"), c)
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}
	page, pageSize, isPaginationValid := utils.ParsePagination(c, defaultPlaylistItemsPageSize)
	if !isPaginationValid {
		return
	}
	clientID, isClientIDValid := clientidentity.Resolve(c)
	if !isClientIDValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	itemsPage, err := h.service.GetPlaylistItemsPage(clientID, id, page, pageSize)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_VIDEO_NOT_FOUND")})
		return
	}
	c.JSON(http.StatusOK, itemsPage)
}

func (h *Handler) SetPlaylistHiddenHandler(c *gin.Context) {
	id := utils.ParseInt(c.Param("id"), c)
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}
	var req SetPlaylistHiddenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}
	if err := h.service.SetPlaylistHidden(id, req.Hidden); err != nil {
		respondVideoError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) AddPlaylistVideoHandler(c *gin.Context) {
	id := utils.ParseInt(c.Param("id"), c)
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}
	var req AddPlaylistVideoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}
	if err := h.service.AddVideoToPlaylist(id, req.VideoID); err != nil {
		respondVideoError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true})
}

func (h *Handler) RemovePlaylistVideoHandler(c *gin.Context) {
	id := utils.ParseInt(c.Param("id"), c)
	videoID := utils.ParseInt(c.Param("videoId"), c)
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}
	if err := h.service.RemoveVideoFromPlaylist(id, videoID); err != nil {
		respondVideoError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) UpdatePlaylistHandler(c *gin.Context) {
	id := utils.ParseInt(c.Param("id"), c)
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}
	var req UpdatePlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}
	if err := h.service.UpdatePlaylistName(id, req.Name); err != nil {
		respondVideoError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) ReorderPlaylistHandler(c *gin.Context) {
	id := utils.ParseInt(c.Param("id"), c)
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}
	var req ReorderPlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}
	if err := h.service.ReorderPlaylistItems(id, req.Items); err != nil {
		respondVideoError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) TrackBehaviorEventHandler(c *gin.Context) {
	var req TrackBehaviorEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	clientID, isClientIDValid := clientidentity.Resolve(c)
	if !isClientIDValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	if err := h.service.TrackBehaviorEvent(clientID, req); err != nil {
		respondVideoError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true})
}

func (h *Handler) GetUnassignedVideosHandler(c *gin.Context) {
	limit := utils.ParseInt(c.DefaultQuery("limit", "2000"), c)
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}
	videos, err := h.service.GetUnassignedVideos(limit)
	if err != nil {
		respondVideoError(c, err)
		return
	}
	c.JSON(http.StatusOK, videos)
}

func (h *Handler) ListLibraryVideosHandler(c *gin.Context) {
	page, pageSize, isPaginationValid := utils.ParsePagination(c, 50)
	if !isPaginationValid {
		return
	}
	query := c.DefaultQuery("query", "")

	videos, err := h.service.ListLibraryVideos(page, pageSize, query)
	if err != nil {
		respondVideoError(c, err)
		return
	}

	c.JSON(http.StatusOK, videos)
}
