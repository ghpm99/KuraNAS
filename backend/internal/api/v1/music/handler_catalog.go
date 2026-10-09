package music

import (
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"
	"net/http"

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

	clientID, isClientIDValid := resolvePlayerClientID(c)
	if !isClientIDValid {
		respondInvalidRequest(c)
		return
	}
	playlists, err := handler.service.GetAutomaticPlaylists(clientID)
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
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}
	sort, isSortValid := parseCatalogSortQuery(c, false)
	if !isSortValid {
		return
	}
	clientID, isClientIDValid := resolvePlayerClientID(c)
	if !isClientIDValid {
		respondInvalidRequest(c)
		return
	}
	catalog, err := handler.service.GetHomeCatalog(clientID, limit, sort)
	if err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		respondMusicError(c, err)
		return
	}

	handler.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, catalog)
}

func (handler *Handler) GetLibraryTracksHandler(c *gin.Context) {
	page, pageSize, isPaginationValid := utils.ParsePagination(c, 50)
	if !isPaginationValid {
		return
	}
	handler.respondLibraryTracks(c, "GetMusicLibraryTracks", "Fetching music library tracks", func() (any, error) {
		return handler.service.GetLibraryTracks(page, pageSize)
	})
}

func (handler *Handler) SearchLibraryTracksHandler(c *gin.Context) {
	page, pageSize, isPaginationValid := utils.ParsePagination(c, 50)
	if !isPaginationValid {
		return
	}
	searchText := c.Query("q")
	handler.respondLibraryTracks(c, "SearchMusicLibraryTracks", "Searching music library tracks", func() (any, error) {
		return handler.service.SearchLibraryTracks(searchText, page, pageSize)
	})
}

func (handler *Handler) GetLibraryArtistsHandler(c *gin.Context) {
	page, pageSize, isPaginationValid := utils.ParsePagination(c, 50)
	if !isPaginationValid {
		return
	}
	sort, isSortValid := parseCatalogSortQuery(c, false)
	if !isSortValid {
		return
	}
	handler.respondLibraryTracks(c, "GetMusicLibraryArtists", "Fetching music artists catalog", func() (any, error) {
		return handler.service.GetLibraryArtists(page, pageSize, sort)
	})
}

func (handler *Handler) GetLibraryTracksByArtistHandler(c *gin.Context) {
	page, pageSize, isPaginationValid := utils.ParsePagination(c, 50)
	if !isPaginationValid {
		return
	}
	artistKey := c.Param("key")
	handler.respondLibraryTracks(c, "GetMusicTracksByArtist", "Fetching music tracks by artist", func() (any, error) {
		return handler.service.GetLibraryTracksByArtist(artistKey, page, pageSize)
	})
}

func (handler *Handler) GetLibraryAlbumsHandler(c *gin.Context) {
	page, pageSize, isPaginationValid := utils.ParsePagination(c, 50)
	if !isPaginationValid {
		return
	}
	sort, isSortValid := parseCatalogSortQuery(c, true)
	if !isSortValid {
		return
	}
	handler.respondLibraryTracks(c, "GetMusicLibraryAlbums", "Fetching music albums catalog", func() (any, error) {
		return handler.service.GetLibraryAlbums(page, pageSize, sort)
	})
}

func (handler *Handler) GetLibraryTracksByAlbumHandler(c *gin.Context) {
	page, pageSize, isPaginationValid := utils.ParsePagination(c, 50)
	if !isPaginationValid {
		return
	}
	albumKey := c.Param("key")
	handler.respondLibraryTracks(c, "GetMusicTracksByAlbum", "Fetching music tracks by album", func() (any, error) {
		return handler.service.GetLibraryTracksByAlbum(albumKey, page, pageSize)
	})
}

func (handler *Handler) GetLibraryGenresHandler(c *gin.Context) {
	page, pageSize, isPaginationValid := utils.ParsePagination(c, 50)
	if !isPaginationValid {
		return
	}
	sort, isSortValid := parseCatalogSortQuery(c, false)
	if !isSortValid {
		return
	}
	handler.respondLibraryTracks(c, "GetMusicLibraryGenres", "Fetching music genres catalog", func() (any, error) {
		return handler.service.GetLibraryGenres(page, pageSize, sort)
	})
}

func (handler *Handler) GetLibraryTracksByGenreHandler(c *gin.Context) {
	page, pageSize, isPaginationValid := utils.ParsePagination(c, 50)
	if !isPaginationValid {
		return
	}
	genreKey := c.Param("key")
	handler.respondLibraryTracks(c, "GetMusicTracksByGenre", "Fetching music tracks by genre", func() (any, error) {
		return handler.service.GetLibraryTracksByGenre(genreKey, page, pageSize)
	})
}

func (handler *Handler) GetLibraryFoldersHandler(c *gin.Context) {
	page, pageSize, isPaginationValid := utils.ParsePagination(c, 50)
	if !isPaginationValid {
		return
	}
	sort, isSortValid := parseCatalogSortQuery(c, false)
	if !isSortValid {
		return
	}
	handler.respondLibraryTracks(c, "GetMusicLibraryFolders", "Fetching music folders catalog", func() (any, error) {
		return handler.service.GetLibraryFolders(page, pageSize, sort)
	})
}

func (handler *Handler) GetLibraryTracksByFolderHandler(c *gin.Context) {
	page, pageSize, isPaginationValid := utils.ParsePagination(c, 50)
	if !isPaginationValid {
		return
	}
	folderKey := c.Param("key")
	handler.respondLibraryTracks(c, "GetMusicTracksByFolder", "Fetching music tracks by folder", func() (any, error) {
		return handler.service.GetLibraryTracksByFolder(folderKey, page, pageSize)
	})
}

func parseCatalogSortQuery(c *gin.Context, isYearAllowed bool) (CatalogSort, bool) {
	sort, isValid := ParseCatalogSort(c.Query("sort"), c.Query("order"), isYearAllowed)
	if !isValid {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_CATALOG_SORT")})
		return CatalogSort{}, false
	}
	return sort, true
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

func (handler *Handler) GetLibraryAlbumSummaryHandler(c *gin.Context) {
	albumKey := c.Param("key")
	handler.respondLibraryTracks(c, "GetMusicAlbumSummary", "Fetching music album summary", func() (any, error) {
		return handler.service.GetLibraryAlbumSummary(albumKey)
	})
}

func (handler *Handler) GetLibraryArtistSummaryHandler(c *gin.Context) {
	artistKey := c.Param("key")
	handler.respondLibraryTracks(c, "GetMusicArtistSummary", "Fetching music artist summary", func() (any, error) {
		return handler.service.GetLibraryArtistSummary(artistKey)
	})
}

func (handler *Handler) GetLibraryGenreSummaryHandler(c *gin.Context) {
	genreKey := c.Param("key")
	handler.respondLibraryTracks(c, "GetMusicGenreSummary", "Fetching music genre summary", func() (any, error) {
		return handler.service.GetLibraryGenreSummary(genreKey)
	})
}

func (handler *Handler) GetLibraryFolderSummaryHandler(c *gin.Context) {
	folderKey := c.Param("key")
	handler.respondLibraryTracks(c, "GetMusicFolderSummary", "Fetching music folder summary", func() (any, error) {
		return handler.service.GetLibraryFolderSummary(folderKey)
	})
}

func (handler *Handler) GetLibraryAlbumsByArtistHandler(c *gin.Context) {
	page, pageSize, isPaginationValid := utils.ParsePagination(c, 50)
	if !isPaginationValid {
		return
	}
	artistKey := c.Param("key")
	handler.respondLibraryTracks(c, "GetMusicAlbumsByArtist", "Fetching music albums by artist", func() (any, error) {
		return handler.service.GetLibraryAlbumsByArtist(artistKey, page, pageSize)
	})
}
