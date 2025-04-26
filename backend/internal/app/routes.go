package app

import (
	"nas-go/api/internal/api/v1/configuration"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, context *AppContext) {

	registerCorsRoutes(router)
	routesV1 := router.Group("/api/v1")
	RegisterFilesRoutes(routesV1, context)
	RegisterConfigRoutes(routesV1, context)
	registerReactRoutes(router)
}

func RegisterFilesRoutes(router *gin.RouterGroup, context *AppContext) {

	files := router.Group("/files")

	files.GET("/", context.Files.Handler.GetFilesHandler)
	files.GET("/path", context.Files.Handler.GetFilesByPathHandler)
	files.GET("/path/:path", context.Files.Handler.GetFilesByPathHandler)
	files.POST("/update", context.Files.Handler.UpdateFilesHandler)

}

func RegisterConfigRoutes(router *gin.RouterGroup, context *AppContext) {
	configurations := router.Group("/configuration")

	configurations.GET("/translation", configuration.GetTranslationJson)
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
		AllowMethods:     []string{"GET", "PUT", "PATCH"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "https://github.com"
		},
		MaxAge: 12 * time.Hour,
	}))
}
