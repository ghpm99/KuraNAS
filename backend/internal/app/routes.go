package app

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, context *AppContext) {

	registerCorsRoutes(router)
	routesV1 := router.Group("/api/v1")
	RegisterFilesRoutes(routesV1, context)
	RegisterDiaryRoutes(routesV1, context)
	RegisterConfigRoutes(routesV1, context)
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
	files.GET("/blob/:id", context.Files.Handler.GetBlobFileHandler)
	files.POST("/update", context.Files.Handler.UpdateFilesHandler)
	files.POST("/starred/:id", context.Files.Handler.StarreFileHandler)
	files.GET("/total-space-used", context.Files.Handler.GetTotalSpaceUsedHandler)
	files.GET("/total-files", context.Files.Handler.GetTotalFilesHandler)
	files.GET("/total-directory", context.Files.Handler.GetTotalDirectoryHandler)
	files.GET("/report-size-by-format", context.Files.Handler.GetReportSizeByFormatHandler)
	files.GET("/top-files-by-size", context.Files.Handler.GetTopFilesBySizeHandler)
	files.GET("/duplicate-files", context.Files.Handler.GetDuplicateFilesHandler)
}

func RegisterDiaryRoutes(router *gin.RouterGroup, context *AppContext) {

	diaryGroup := router.Group("/diary")

	diaryGroup.GET("/", context.Diary.Handler.GetDiaryHandler)
	diaryGroup.GET("/summary", context.Diary.Handler.GetSummaryHandler)
	diaryGroup.POST("/", context.Diary.Handler.CreateDiaryHandler)
	diaryGroup.PUT("/:id", context.Diary.Handler.UpdateDiaryHandler)
	diaryGroup.POST("/copy", context.Diary.Handler.DuplicateDiaryHandler)
}

func RegisterConfigRoutes(router *gin.RouterGroup, context *AppContext) {
	configurations := router.Group("/configuration")

	configurations.GET("/translation", context.ConfigurationHandler.GetTranslationJson)
	configurations.GET("/about", context.ConfigurationHandler.GetAboutHandler)
}

func registerReactRoutes(router *gin.Engine) {
	router.Static("/assets", "./dist/assets")

	router.NoRoute(func(c *gin.Context) {
		c.File("./dist/index.html")
	})

	router.Static("/frontend", "/dist")
}

func registerCorsRoutes(router *gin.Engine) {
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "PUT", "POST"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "https://github.com"
		},
		MaxAge: 12 * time.Hour,
	}))
}
