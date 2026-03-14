package music

import (
	"net/http"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

func (handler *Handler) GetAutomaticPlaylistsHandler(c *gin.Context) {
	loggerModel, _ := handler.logService.CreateLog(logger.LoggerModel{
		Name:        "GetAutomaticPlaylists",
		Description: "Fetching automatic music playlists",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	playlists, err := handler.service.GetAutomaticPlaylists(c.ClientIP())
	if err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		respondMusicError(c, err)
		return
	}

	handler.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, playlists)
}

func (handler *Handler) GetHomeCatalogHandler(c *gin.Context) {
	loggerModel, _ := handler.logService.CreateLog(logger.LoggerModel{
		Name:        "GetMusicHomeCatalog",
		Description: "Fetching music home catalog",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	limit := utils.ParseInt(c.DefaultQuery("limit", "4"), c)
	catalog, err := handler.service.GetHomeCatalog(c.ClientIP(), limit)
	if err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		respondMusicError(c, err)
		return
	}

	handler.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, catalog)
}

func (handler *Handler) GetLibraryTracksHandler(c *gin.Context) {
	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "50"), c)
	handler.respondLibraryTracks(c, "GetMusicLibraryTracks", "Fetching music library tracks", func() (any, error) {
		return handler.service.GetLibraryTracks(page, pageSize)
	})
}

func (handler *Handler) GetLibraryArtistsHandler(c *gin.Context) {
	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "50"), c)
	handler.respondLibraryTracks(c, "GetMusicLibraryArtists", "Fetching music artists catalog", func() (any, error) {
		return handler.service.GetLibraryArtists(page, pageSize)
	})
}

func (handler *Handler) GetLibraryTracksByArtistHandler(c *gin.Context) {
	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "50"), c)
	artistKey := c.Param("key")
	handler.respondLibraryTracks(c, "GetMusicTracksByArtist", "Fetching music tracks by artist", func() (any, error) {
		return handler.service.GetLibraryTracksByArtist(artistKey, page, pageSize)
	})
}

func (handler *Handler) GetLibraryAlbumsHandler(c *gin.Context) {
	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "50"), c)
	handler.respondLibraryTracks(c, "GetMusicLibraryAlbums", "Fetching music albums catalog", func() (any, error) {
		return handler.service.GetLibraryAlbums(page, pageSize)
	})
}

func (handler *Handler) GetLibraryTracksByAlbumHandler(c *gin.Context) {
	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "50"), c)
	albumKey := c.Param("key")
	handler.respondLibraryTracks(c, "GetMusicTracksByAlbum", "Fetching music tracks by album", func() (any, error) {
		return handler.service.GetLibraryTracksByAlbum(albumKey, page, pageSize)
	})
}

func (handler *Handler) GetLibraryGenresHandler(c *gin.Context) {
	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "50"), c)
	handler.respondLibraryTracks(c, "GetMusicLibraryGenres", "Fetching music genres catalog", func() (any, error) {
		return handler.service.GetLibraryGenres(page, pageSize)
	})
}

func (handler *Handler) GetLibraryTracksByGenreHandler(c *gin.Context) {
	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "50"), c)
	genreKey := c.Param("key")
	handler.respondLibraryTracks(c, "GetMusicTracksByGenre", "Fetching music tracks by genre", func() (any, error) {
		return handler.service.GetLibraryTracksByGenre(genreKey, page, pageSize)
	})
}

func (handler *Handler) GetLibraryFoldersHandler(c *gin.Context) {
	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "50"), c)
	handler.respondLibraryTracks(c, "GetMusicLibraryFolders", "Fetching music folders catalog", func() (any, error) {
		return handler.service.GetLibraryFolders(page, pageSize)
	})
}

func (handler *Handler) respondLibraryTracks(c *gin.Context, name string, description string, run func() (any, error)) {
	loggerModel, _ := handler.logService.CreateLog(logger.LoggerModel{
		Name:        name,
		Description: description,
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	payload, err := run()
	if err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		respondMusicError(c, err)
		return
	}

	handler.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, payload)
}
