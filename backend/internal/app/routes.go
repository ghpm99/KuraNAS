package app

import (
	"nas-go/api/internal/api/v1/health"
	"nas-go/api/internal/config"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, context *AppContext) {

	registerCorsRoutes(router, context)
	router.Use(gzip.Gzip(gzip.DefaultCompression))
	registerSwaggerRoutes(router)
	routesV1 := router.Group("/api/v1")
	RegisterFilesRoutes(routesV1, context)
	RegisterDiaryRoutes(routesV1, context)
	RegisterMusicRoutes(routesV1, context)
	RegisterVideoRoutes(routesV1, context)
	RegisterAnalyticsRoutes(routesV1, context)
	RegisterJobsRoutes(routesV1, context)
	RegisterConfigRoutes(routesV1, context)
	RegisterUpdateRoutes(routesV1, context)
	RegisterSearchRoutes(routesV1, context)
	RegisterNotificationRoutes(routesV1, context)
	RegisterHealthRoutes(routesV1)
	registerReactRoutes(router)
}

func RegisterFilesRoutes(router *gin.RouterGroup, context *AppContext) {

	files := router.Group("/files")

	files.GET("/", context.Files.Handler.GetFilesHandler)
	files.GET("/tree", context.Files.Handler.GetFilesTreeHandler)
	files.GET("/:id", context.Files.Handler.GetChildrenByIdHandler)
	files.GET("/recent", context.Files.Handler.GetRecentFilesHandler)
	files.GET("/recent/:id", context.Files.Handler.GetRecentAccessByFileHandler)
	files.GET("/path", context.Files.Handler.GetFilesByPathHandler)
	files.GET("/path/:path", context.Files.Handler.GetFilesByPathHandler)
	files.GET("/thumbnail/:id", context.Files.Handler.GetFileThumbnailHandler)
	files.GET("/video-thumbnail/:id", context.Files.Handler.GetVideoThumbnailHandler)
	files.GET("/video-preview/:id", context.Files.Handler.GetVideoPreviewHandler)
	files.GET("/blob/:id", context.Files.Handler.GetBlobFileHandler)
	files.POST("/update", context.Files.Handler.UpdateFilesHandler)
	files.POST("/upload", context.Files.Handler.UploadFilesHandler)
	files.POST("/folder", context.Files.Handler.CreateFolderHandler)
	files.POST("/move", context.Files.Handler.MoveFileHandler)
	files.POST("/copy", context.Files.Handler.CopyFileHandler)
	files.POST("/rename", context.Files.Handler.RenameFileHandler)
	files.DELETE("/path", context.Files.Handler.DeleteFileHandler)
	files.POST("/starred/:id", context.Files.Handler.StarreFileHandler)
	files.GET("/total-space-used", context.Files.Handler.GetTotalSpaceUsedHandler)
	files.GET("/total-files", context.Files.Handler.GetTotalFilesHandler)
	files.GET("/total-directory", context.Files.Handler.GetTotalDirectoryHandler)
	files.GET("/report-size-by-format", context.Files.Handler.GetReportSizeByFormatHandler)
	files.GET("/top-files-by-size", context.Files.Handler.GetTopFilesBySizeHandler)
	files.GET("/duplicate-files", context.Files.Handler.GetDuplicateFilesHandler)
	files.GET("/images", context.Files.Handler.GetImagesHandler)
	files.GET("/music", context.Files.Handler.GetMusicHandler)
	files.GET("/videos", context.Files.Handler.GetVideosHandler)
	files.GET("/stream/:id", context.Files.Handler.StreamAudioHandler)
	files.GET("/video-stream/:id", context.Files.Handler.StreamVideoHandler)

	music := files.Group("/music")
	music.GET("/artists", context.Files.Handler.GetMusicArtistsHandler)
	music.GET("/artists/:name", context.Files.Handler.GetMusicByArtistHandler)
	music.GET("/albums", context.Files.Handler.GetMusicAlbumsHandler)
	music.GET("/albums/:name", context.Files.Handler.GetMusicByAlbumHandler)
	music.GET("/genres", context.Files.Handler.GetMusicGenresHandler)
	music.GET("/genres/:name", context.Files.Handler.GetMusicByGenreHandler)
	music.GET("/folders", context.Files.Handler.GetMusicFoldersHandler)
}

func RegisterDiaryRoutes(router *gin.RouterGroup, context *AppContext) {

	diaryGroup := router.Group("/diary")

	diaryGroup.GET("/", context.Diary.Handler.GetDiaryHandler)
	diaryGroup.GET("/summary", context.Diary.Handler.GetSummaryHandler)
	diaryGroup.POST("/", context.Diary.Handler.CreateDiaryHandler)
	diaryGroup.PUT("/:id", context.Diary.Handler.UpdateDiaryHandler)
	diaryGroup.POST("/copy", context.Diary.Handler.DuplicateDiaryHandler)
}

func RegisterMusicRoutes(router *gin.RouterGroup, context *AppContext) {
	playlists := router.Group("/music/playlists")
	library := router.Group("/music/library")

	playlists.GET("/", context.Music.Handler.GetPlaylistsHandler)
	playlists.POST("/", context.Music.Handler.CreatePlaylistHandler)
	playlists.GET("/now-playing", context.Music.Handler.GetNowPlayingHandler)
	playlists.GET("/system", context.Music.Handler.GetAutomaticPlaylistsHandler)
	playlists.GET("/:id", context.Music.Handler.GetPlaylistByIDHandler)
	playlists.PUT("/:id", context.Music.Handler.UpdatePlaylistHandler)
	playlists.DELETE("/:id", context.Music.Handler.DeletePlaylistHandler)
	playlists.GET("/:id/tracks", context.Music.Handler.GetPlaylistTracksHandler)
	playlists.POST("/:id/tracks", context.Music.Handler.AddPlaylistTrackHandler)
	playlists.DELETE("/:id/tracks/:fileId", context.Music.Handler.RemovePlaylistTrackHandler)
	playlists.PUT("/:id/tracks/reorder", context.Music.Handler.ReorderPlaylistTracksHandler)

	library.GET("", context.Music.Handler.GetLibraryTracksHandler)
	library.GET("/", context.Music.Handler.GetLibraryTracksHandler)
	library.GET("/home", context.Music.Handler.GetHomeCatalogHandler)
	library.GET("/artists", context.Music.Handler.GetLibraryArtistsHandler)
	library.GET("/artists/:key/tracks", context.Music.Handler.GetLibraryTracksByArtistHandler)
	library.GET("/albums", context.Music.Handler.GetLibraryAlbumsHandler)
	library.GET("/albums/:key/tracks", context.Music.Handler.GetLibraryTracksByAlbumHandler)
	library.GET("/genres", context.Music.Handler.GetLibraryGenresHandler)
	library.GET("/genres/:key/tracks", context.Music.Handler.GetLibraryTracksByGenreHandler)
	library.GET("/folders", context.Music.Handler.GetLibraryFoldersHandler)
	library.GET("/folders/:key/tracks", context.Music.Handler.GetLibraryTracksByFolderHandler)

	playerState := router.Group("/music/player-state")
	playerState.GET("/", context.Music.Handler.GetPlayerStateHandler)
	playerState.PUT("/", context.Music.Handler.UpdatePlayerStateHandler)
}

func RegisterConfigRoutes(router *gin.RouterGroup, context *AppContext) {
	configurations := router.Group("/configuration")

	configurations.GET("/translation", context.Configuration.Handler.GetTranslationJson)
	configurations.GET("/about", context.Configuration.Handler.GetAboutHandler)
	configurations.GET("/settings", context.Configuration.Handler.GetSettingsHandler)
	configurations.PUT("/settings", context.Configuration.Handler.UpdateSettingsHandler)
}

func RegisterVideoRoutes(router *gin.RouterGroup, context *AppContext) {
	playback := router.Group("/video/playback")
	catalog := router.Group("/video/catalog")
	library := router.Group("/video/library")
	playlists := router.Group("/video/playlists")

	playback.POST("/start", context.Video.Handler.StartPlaybackHandler)
	playback.GET("/state", context.Video.Handler.GetPlaybackStateHandler)
	playback.PUT("/state", context.Video.Handler.UpdatePlaybackStateHandler)
	playback.POST("/next", context.Video.Handler.NextVideoHandler)
	playback.POST("/previous", context.Video.Handler.PreviousVideoHandler)
	playback.POST("/behavior", context.Video.Handler.TrackBehaviorEventHandler)
	catalog.GET("/home", context.Video.Handler.GetHomeCatalogHandler)
	library.GET("/files", context.Video.Handler.ListLibraryVideosHandler)

	playlists.GET("/", context.Video.Handler.GetPlaylistsHandler)
	playlists.GET("", context.Video.Handler.GetPlaylistsHandler)
	playlists.GET("/memberships", context.Video.Handler.GetPlaylistMembershipsHandler)
	playlists.POST("/rebuild", context.Video.Handler.RebuildPlaylistsHandler)
	playlists.GET("/unassigned", context.Video.Handler.GetUnassignedVideosHandler)
	playlists.PUT("/:id/reorder", context.Video.Handler.ReorderPlaylistHandler)
	playlists.GET("/:id", context.Video.Handler.GetPlaylistByIDHandler)
	playlists.PUT("/:id", context.Video.Handler.UpdatePlaylistHandler)
	playlists.PUT("/:id/hidden", context.Video.Handler.SetPlaylistHiddenHandler)
	playlists.POST("/:id/videos", context.Video.Handler.AddPlaylistVideoHandler)
	playlists.DELETE("/:id/videos/:videoId", context.Video.Handler.RemovePlaylistVideoHandler)
}

func RegisterUpdateRoutes(router *gin.RouterGroup, context *AppContext) {
	update := router.Group("/update")

	update.GET("/status", context.UpdateHandler.GetUpdateStatusHandler)
	update.POST("/apply", context.UpdateHandler.ApplyUpdateHandler)
}

func RegisterSearchRoutes(router *gin.RouterGroup, context *AppContext) {
	if context == nil || context.Search == nil || context.Search.Handler == nil {
		return
	}

	search := router.Group("/search")
	search.GET("/global", context.Search.Handler.SearchGlobalHandler)
}

func RegisterAnalyticsRoutes(router *gin.RouterGroup, context *AppContext) {
	analytics := router.Group("/analytics")
	analytics.GET("/overview", context.Analytics.Handler.GetOverviewHandler)
}

func RegisterJobsRoutes(router *gin.RouterGroup, context *AppContext) {
	if context == nil || context.Jobs == nil || context.Jobs.Handler == nil {
		return
	}

	jobs := router.Group("/jobs")
	jobs.GET("/:id", context.Jobs.Handler.GetJobByIDHandler)
	jobs.GET("", context.Jobs.Handler.ListJobsHandler)
	jobs.GET("/:id/steps", context.Jobs.Handler.GetStepsByJobIDHandler)
	jobs.POST("/:id/cancel", context.Jobs.Handler.CancelJobHandler)
}

func RegisterNotificationRoutes(router *gin.RouterGroup, context *AppContext) {
	if context == nil || context.Notifications == nil || context.Notifications.Handler == nil {
		return
	}

	notifs := router.Group("/notifications")
	notifs.GET("", context.Notifications.Handler.ListNotificationsHandler)
	notifs.GET("/unread-count", context.Notifications.Handler.GetUnreadCountHandler)
	notifs.GET("/:id", context.Notifications.Handler.GetNotificationByIDHandler)
	notifs.PUT("/:id/read", context.Notifications.Handler.MarkAsReadHandler)
	notifs.PUT("/read-all", context.Notifications.Handler.MarkAllAsReadHandler)
}

func RegisterHealthRoutes(router *gin.RouterGroup) {
	healthHandler := health.NewHandler()
	router.GET("/health", healthHandler.GetHealthHandler)
}

func registerReactRoutes(router *gin.Engine) {
	router.Group("/assets", cacheControlMiddleware("public, max-age=31536000, immutable")).
		Static("/", "./dist/assets")

	router.NoRoute(func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache")
		c.File("./dist/index.html")
	})
}

func cacheControlMiddleware(value string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", value)
		c.Next()
	}
}

func registerCorsRoutes(router *gin.Engine, context *AppContext) {
	// Get allowed origins from environment variable (comma-separated)
	// Default to localhost for development
	allowedOriginsStr := config.AppConfig.AllowedOrigins
	allowedOrigins := strings.Split(allowedOriginsStr, ",")

	// Trim whitespace from each origin
	for i, origin := range allowedOrigins {
		allowedOrigins[i] = strings.TrimSpace(origin)
	}

	router.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "PUT", "POST", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
}
