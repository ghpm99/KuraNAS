package music

import (
	"database/sql"
	"errors"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service    ServiceInterface
	logService logger.LoggerServiceInterface
}

func NewHandler(musicService ServiceInterface, loggerService logger.LoggerServiceInterface) *Handler {
	return &Handler{
		service:    musicService,
		logService: loggerService,
	}
}

func respondMusicError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_MUSIC_NOT_FOUND")})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_MUSIC_OPERATION_FAILED")})
	}
}

func (handler *Handler) GetPlaylistsHandler(c *gin.Context) {
	loggerModel, _ := handler.logService.CreateLog(logger.LoggerModel{
		Name:        "GetPlaylists",
		Description: "Fetching playlists",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "50"), c)

	pagination, err := handler.service.GetPlaylists(page, pageSize)
	if err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		respondMusicError(c, err)
		return
	}

	handler.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, pagination)
}

func (handler *Handler) GetPlaylistByIDHandler(c *gin.Context) {
	loggerModel, _ := handler.logService.CreateLog(logger.LoggerModel{
		Name:        "GetPlaylistByID",
		Description: "Fetching playlist by ID",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	id := utils.ParseInt(c.Param("id"), c)

	playlist, err := handler.service.GetPlaylistByID(id)
	if err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_MUSIC_NOT_FOUND")})
		return
	}

	handler.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, playlist)
}

func (handler *Handler) CreatePlaylistHandler(c *gin.Context) {
	loggerModel, _ := handler.logService.CreateLog(logger.LoggerModel{
		Name:        "CreatePlaylist",
		Description: "Creating a new playlist",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	var req CreatePlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	playlist, err := handler.service.CreatePlaylist(req)
	if err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		respondMusicError(c, err)
		return
	}

	handler.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusCreated, playlist)
}

func (handler *Handler) UpdatePlaylistHandler(c *gin.Context) {
	loggerModel, _ := handler.logService.CreateLog(logger.LoggerModel{
		Name:        "UpdatePlaylist",
		Description: "Updating a playlist",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	id := utils.ParseInt(c.Param("id"), c)

	var req UpdatePlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	playlist, err := handler.service.UpdatePlaylist(id, req)
	if err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		respondMusicError(c, err)
		return
	}

	handler.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, playlist)
}

func (handler *Handler) DeletePlaylistHandler(c *gin.Context) {
	loggerModel, _ := handler.logService.CreateLog(logger.LoggerModel{
		Name:        "DeletePlaylist",
		Description: "Deleting a playlist",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	id := utils.ParseInt(c.Param("id"), c)

	err := handler.service.DeletePlaylist(id)
	if err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		respondMusicError(c, err)
		return
	}

	handler.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (handler *Handler) GetPlaylistTracksHandler(c *gin.Context) {
	loggerModel, _ := handler.logService.CreateLog(logger.LoggerModel{
		Name:        "GetPlaylistTracks",
		Description: "Fetching playlist tracks",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	id := utils.ParseInt(c.Param("id"), c)
	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "50"), c)

	pagination, err := handler.service.GetPlaylistTracks(id, page, pageSize)
	if err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		respondMusicError(c, err)
		return
	}

	handler.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, pagination)
}

func (handler *Handler) AddPlaylistTrackHandler(c *gin.Context) {
	loggerModel, _ := handler.logService.CreateLog(logger.LoggerModel{
		Name:        "AddPlaylistTrack",
		Description: "Adding track to playlist",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	id := utils.ParseInt(c.Param("id"), c)

	var req AddTrackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	track, err := handler.service.AddPlaylistTrack(id, req.FileID)
	if err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		respondMusicError(c, err)
		return
	}

	handler.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusCreated, track)
}

func (handler *Handler) RemovePlaylistTrackHandler(c *gin.Context) {
	loggerModel, _ := handler.logService.CreateLog(logger.LoggerModel{
		Name:        "RemovePlaylistTrack",
		Description: "Removing track from playlist",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	id := utils.ParseInt(c.Param("id"), c)
	fileId := utils.ParseInt(c.Param("fileId"), c)

	err := handler.service.RemovePlaylistTrack(id, fileId)
	if err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		respondMusicError(c, err)
		return
	}

	handler.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (handler *Handler) ReorderPlaylistTracksHandler(c *gin.Context) {
	loggerModel, _ := handler.logService.CreateLog(logger.LoggerModel{
		Name:        "ReorderPlaylistTracks",
		Description: "Reordering playlist tracks",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	id := utils.ParseInt(c.Param("id"), c)

	var req ReorderTrackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	err := handler.service.ReorderPlaylistTracks(id, req.Tracks)
	if err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		respondMusicError(c, err)
		return
	}

	handler.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (handler *Handler) GetNowPlayingHandler(c *gin.Context) {
	loggerModel, _ := handler.logService.CreateLog(logger.LoggerModel{
		Name:        "GetNowPlaying",
		Description: "Fetching now playing queue",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	playlist, err := handler.service.GetOrCreateNowPlaying()
	if err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		respondMusicError(c, err)
		return
	}

	handler.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, playlist)
}

func (handler *Handler) GetPlayerStateHandler(c *gin.Context) {
	loggerModel, _ := handler.logService.CreateLog(logger.LoggerModel{
		Name:        "GetPlayerState",
		Description: "Fetching player state",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	clientID := c.ClientIP()

	state, err := handler.service.GetPlayerState(clientID)
	if err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_MUSIC_NOT_FOUND")})
		return
	}

	handler.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, state)
}

func (handler *Handler) UpdatePlayerStateHandler(c *gin.Context) {
	loggerModel, _ := handler.logService.CreateLog(logger.LoggerModel{
		Name:        "UpdatePlayerState",
		Description: "Updating player state",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	clientID := c.ClientIP()

	var req UpdatePlayerStateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	state, err := handler.service.UpdatePlayerState(clientID, req)
	if err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		respondMusicError(c, err)
		return
	}

	handler.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, state)
}
