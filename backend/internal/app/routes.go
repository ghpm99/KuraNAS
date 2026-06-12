package app

import (
	"nas-go/api/internal/api/v1/accesscontrol"
	"nas-go/api/internal/api/v1/email"
	"nas-go/api/internal/api/v1/health"
	"nas-go/api/internal/config"
	"nas-go/api/internal/dav"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, context *AppContext) {

	registerAccessControlMiddleware(router, context)
	registerCorsRoutes(router, context)
	// WebDAV registers before the gzip middleware on purpose: compressing
	// PUT/PROPFIND bodies corrupts them for native clients. It still sits
	// behind the IP whitelist installed above.
	registerWebDAVRoutes(router)
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
	RegisterCapturesRoutes(routesV1, context)
	RegisterLibrariesRoutes(routesV1, context)
	RegisterAIProvidersRoutes(routesV1, context)
	RegisterOllamaRoutes(routesV1, context)
	RegisterAssistantRoutes(routesV1, context)
	RegisterWatchFoldersRoutes(routesV1, context)
	RegisterTakeoutRoutes(routesV1, context)
	RegisterDistributionRoutes(routesV1, context)
	RegisterHealthRoutes(routesV1)
	RegisterAccessControlRoutes(routesV1, context)
	RegisterTrashRoutes(routesV1, context)
	RegisterStorageRootsRoutes(routesV1, context)
	RegisterEmailRoutes(routesV1, context)
	registerReactRoutes(router)
}

// registerWebDAVRoutes mounts the WebDAV tree under /dav when enabled
// (WEBDAV_ENABLED env, default off). With the flag off the route simply does
// not exist.
func registerWebDAVRoutes(router *gin.Engine) {
	if !config.AppConfig.EnableWebDAV {
		return
	}

	handler := dav.NewHandler()
	router.Any(dav.Prefix+"/*path", gin.WrapH(handler))
	// Mount-point requests arrive without the trailing slash.
	router.Any(dav.Prefix, gin.WrapH(handler))
}

func RegisterStorageRootsRoutes(router *gin.RouterGroup, context *AppContext) {
	if context == nil || context.StorageRoots == nil || context.StorageRoots.Handler == nil {
		return
	}

	group := router.Group("/storage-roots")
	group.GET("", context.StorageRoots.Handler.GetStorageRootsHandler)
	group.POST("", context.StorageRoots.Handler.CreateStorageRootHandler)
	group.PUT("/:id", context.StorageRoots.Handler.UpdateStorageRootHandler)
	group.DELETE("/:id", context.StorageRoots.Handler.DeleteStorageRootHandler)
}

// RegisterEmailRoutes mounts the e-mail accounts feature. When the context is
// nil (no/invalid EMAIL_TOKEN_KEY) the routes still exist but every one of
// them answers an explicit i18n error — the feature refuses to run without
// encryption at rest.
func RegisterEmailRoutes(router *gin.RouterGroup, context *AppContext) {
	group := router.Group("/email")

	if context == nil || context.Email == nil || context.Email.Handler == nil {
		group.Any("/accounts", email.DisabledHandler)
		group.Any("/accounts/*rest", email.DisabledHandler)
		group.Any("/oauth/google/callback", email.DisabledHandler)
		return
	}

	handler := context.Email.Handler
	group.GET("/accounts", handler.GetAccountsHandler)
	group.DELETE("/accounts/:id", handler.DeleteAccountHandler)
	group.PUT("/accounts/:id/sync-enabled", handler.UpdateSyncEnabledHandler)
	group.POST("/accounts/google/auth-url", handler.GoogleAuthURLHandler)
	group.GET("/oauth/google/callback", handler.GoogleCallbackHandler)
	group.POST("/accounts/microsoft/device-code", handler.MicrosoftDeviceCodeHandler)
	group.GET("/accounts/microsoft/device-code/status", handler.MicrosoftDeviceCodeStatusHandler)
}

func RegisterTrashRoutes(router *gin.RouterGroup, context *AppContext) {
	if context == nil || context.Trash == nil || context.Trash.Handler == nil {
		return
	}

	group := router.Group("/trash")
	group.GET("", context.Trash.Handler.GetTrashItemsHandler)
	group.POST("/:id/restore", context.Trash.Handler.RestoreTrashItemHandler)
	group.DELETE("/:id", context.Trash.Handler.DeleteTrashItemHandler)
	group.DELETE("", context.Trash.Handler.EmptyTrashHandler)
	group.GET("/retention", context.Trash.Handler.GetTrashRetentionHandler)
	group.PUT("/retention", context.Trash.Handler.UpdateTrashRetentionHandler)
}

// registerAccessControlMiddleware installs the IP whitelist in front of every
// route — API, assets, SPA and Swagger alike. It runs before CORS so a blocked
// origin gets its 403 straight away.
func registerAccessControlMiddleware(router *gin.Engine, context *AppContext) {
	if context == nil || context.AccessControl == nil || context.AccessControl.Service == nil {
		return
	}
	router.Use(accesscontrol.NewMiddleware(context.AccessControl.Service))
}

func RegisterAccessControlRoutes(router *gin.RouterGroup, context *AppContext) {
	if context == nil || context.AccessControl == nil || context.AccessControl.Handler == nil {
		return
	}

	group := router.Group("/access-control")
	group.GET("/ips", context.AccessControl.Handler.GetAllowedIPsHandler)
	group.POST("/ips", context.AccessControl.Handler.CreateAllowedIPHandler)
	group.PUT("/ips/:id", context.AccessControl.Handler.UpdateAllowedIPHandler)
	group.DELETE("/ips/:id", context.AccessControl.Handler.DeleteAllowedIPHandler)
	group.GET("/client-ip", context.AccessControl.Handler.GetClientIPHandler)
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
	if context.Image != nil {
		files.GET("/images", context.Image.Handler.GetImagesHandler)
	}
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
	// Legacy file-browse routes for music — same paths as before, now served by music handler.
	// These must be registered on the /files group so the paths stay identical for all clients.
	filesGroup := router.Group("/files")
	filesGroup.GET("/music", context.Music.Handler.GetMusicHandler)
	filesGroup.GET("/stream/:id", context.Music.Handler.StreamAudioHandler)
	musicBrowse := filesGroup.Group("/music")
	musicBrowse.GET("/artists", context.Music.Handler.GetMusicArtistsHandler)
	musicBrowse.GET("/artists/:name", context.Music.Handler.GetMusicByArtistHandler)
	musicBrowse.GET("/albums", context.Music.Handler.GetMusicAlbumsHandler)
	musicBrowse.GET("/albums/:name", context.Music.Handler.GetMusicByAlbumHandler)
	musicBrowse.GET("/genres", context.Music.Handler.GetMusicGenresHandler)
	musicBrowse.GET("/genres/:name", context.Music.Handler.GetMusicByGenreHandler)
	musicBrowse.GET("/folders", context.Music.Handler.GetMusicFoldersHandler)

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
	// Browse/streaming endpoints moved from the files domain — paths unchanged
	// (/files/...) to preserve the HTTP contract; only the owner changed.
	filesGroup := router.Group("/files")
	filesGroup.GET("/videos", context.Video.Handler.GetVideosHandler)
	filesGroup.GET("/video-stream/:id", context.Video.Handler.StreamVideoHandler)
	filesGroup.GET("/video-thumbnail/:id", context.Video.Handler.GetVideoThumbnailHandler)
	filesGroup.GET("/video-preview/:id", context.Video.Handler.GetVideoPreviewHandler)

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
	if context == nil || context.Analytics == nil || context.Analytics.Handler == nil {
		return
	}
	handler := context.Analytics.Handler
	analytics := router.Group("/analytics")
	analytics.GET("/storage", handler.GetStorageHandler)
	analytics.GET("/timeseries", handler.GetTimeSeriesHandler)
	analytics.GET("/types", handler.GetTypesHandler)
	analytics.GET("/extensions", handler.GetExtensionsHandler)
	analytics.GET("/recent-files", handler.GetRecentFilesHandler)
	analytics.GET("/top-folders", handler.GetTopFoldersHandler)
	analytics.GET("/hot-folders", handler.GetHotFoldersHandler)
	analytics.GET("/duplicates", handler.GetDuplicatesHandler)
	analytics.GET("/duplicates/groups", handler.GetDuplicateGroupsHandler)
	analytics.GET("/library", handler.GetLibraryHandler)
	analytics.GET("/processing", handler.GetProcessingHandler)
	analytics.GET("/health", handler.GetHealthHandler)
	analytics.GET("/ai-usage", handler.GetAIUsageHandler)
	analytics.GET("/insights", handler.GetInsightsHandler)
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

func RegisterCapturesRoutes(router *gin.RouterGroup, context *AppContext) {
	if context == nil || context.Captures == nil || context.Captures.Handler == nil {
		return
	}

	capturesGroup := router.Group("/captures")
	capturesGroup.POST("/upload", context.Captures.Handler.UploadCaptureHandler)
	capturesGroup.POST("/upload/init", context.Captures.Handler.InitCaptureUploadHandler)
	capturesGroup.POST("/upload/chunk", context.Captures.Handler.UploadCaptureChunkHandler)
	capturesGroup.POST("/upload/complete", context.Captures.Handler.CompleteCaptureUploadHandler)
	capturesGroup.GET("", context.Captures.Handler.GetCapturesHandler)
	capturesGroup.GET("/:id", context.Captures.Handler.GetCaptureByIDHandler)
	capturesGroup.DELETE("/:id", context.Captures.Handler.DeleteCaptureHandler)
}

func RegisterLibrariesRoutes(router *gin.RouterGroup, context *AppContext) {
	if context == nil || context.Libraries == nil || context.Libraries.Handler == nil {
		return
	}

	libraries := router.Group("/libraries")
	libraries.GET("", context.Libraries.Handler.GetLibrariesHandler)
	libraries.GET("/", context.Libraries.Handler.GetLibrariesHandler)
	libraries.PUT("/:category", context.Libraries.Handler.UpdateLibraryHandler)
}

func RegisterAIProvidersRoutes(router *gin.RouterGroup, context *AppContext) {
	if context == nil || context.AIProviders == nil || context.AIProviders.Handler == nil {
		return
	}

	providers := router.Group("/ai/providers")
	providers.GET("", context.AIProviders.Handler.GetProvidersHandler)
	providers.GET("/", context.AIProviders.Handler.GetProvidersHandler)
	providers.PUT("/:name", context.AIProviders.Handler.UpdateProviderHandler)
}

func RegisterOllamaRoutes(router *gin.RouterGroup, context *AppContext) {
	if context == nil || context.Ollama == nil || context.Ollama.Handler == nil {
		return
	}

	ollamaGroup := router.Group("/ai/ollama")
	ollamaGroup.GET("/status", context.Ollama.Handler.GetStatusHandler)
	ollamaGroup.GET("/models", context.Ollama.Handler.ListModelsHandler)
	ollamaGroup.POST("/models/pull", context.Ollama.Handler.PullModelHandler)
	ollamaGroup.DELETE("/models/:name", context.Ollama.Handler.DeleteModelHandler)
}

func RegisterAssistantRoutes(router *gin.RouterGroup, context *AppContext) {
	if context == nil || context.Assistant == nil || context.Assistant.Handler == nil {
		return
	}

	assistantGroup := router.Group("/assistant")
	assistantGroup.POST("/chat", context.Assistant.Handler.ChatHandler)
	assistantGroup.POST("/chat/stream", context.Assistant.Handler.ChatStreamHandler)
	assistantGroup.GET("/conversations", context.Assistant.Handler.ListConversationsHandler)
	assistantGroup.GET("/conversations/:id/messages", context.Assistant.Handler.GetMessagesHandler)
	assistantGroup.DELETE("/conversations/:id", context.Assistant.Handler.DeleteConversationHandler)
}

func RegisterWatchFoldersRoutes(router *gin.RouterGroup, context *AppContext) {
	if context == nil || context.WatchFolders == nil || context.WatchFolders.Handler == nil {
		return
	}

	watchFolders := router.Group("/watch-folders")
	watchFolders.GET("", context.WatchFolders.Handler.GetWatchFoldersHandler)
	watchFolders.GET("/", context.WatchFolders.Handler.GetWatchFoldersHandler)
	watchFolders.POST("", context.WatchFolders.Handler.CreateWatchFolderHandler)
	watchFolders.POST("/", context.WatchFolders.Handler.CreateWatchFolderHandler)
	watchFolders.PUT("/:id", context.WatchFolders.Handler.UpdateWatchFolderHandler)
	watchFolders.DELETE("/:id", context.WatchFolders.Handler.DeleteWatchFolderHandler)
}

func RegisterTakeoutRoutes(router *gin.RouterGroup, context *AppContext) {
	if context == nil || context.Takeout == nil || context.Takeout.Handler == nil {
		return
	}

	takeoutRoutes := router.Group("/takeout")
	takeoutRoutes.POST("/upload/init", context.Takeout.Handler.InitUploadHandler)
	takeoutRoutes.POST("/upload/chunk", context.Takeout.Handler.UploadChunkHandler)
	takeoutRoutes.POST("/upload/complete", context.Takeout.Handler.CompleteUploadHandler)
}

func RegisterDistributionRoutes(router *gin.RouterGroup, context *AppContext) {
	if context == nil || context.Distribution == nil || context.Distribution.Handler == nil {
		return
	}

	downloads := router.Group("/downloads")
	downloads.GET("", context.Distribution.Handler.GetDownloadsHandler)
	downloads.GET("/", context.Distribution.Handler.GetDownloadsHandler)
	downloads.GET("/:id", context.Distribution.Handler.DownloadFileHandler)
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
