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
	registerReactRoutes(router)
}

func RegisterFilesRoutes(router *gin.RouterGroup, context *AppContext) {

	files := router.Group("/files")

	files.GET("/", context.Files.Handler.GetFilesHandler)
	files.POST("/update", context.Files.Handler.UpdateFilesHandler)

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
		AllowOrigins:     []string{"http://localhost:5173"},
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
