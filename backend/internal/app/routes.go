package app

import (
	"database/sql"
	"nas-go/api/internal/api/v1/images"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, database *sql.DB) {

	routesV1 := router.Group("/api/v1")
	images.RegisterImagesRoutes(routesV1, database)
	RegisterReactRoutes(router)
}

func RegisterReactRoutes(router *gin.Engine) {
	router.Static("/assets", "./dist/assets")

	router.NoRoute(func(c *gin.Context) {
		c.File("./dist/index.html")
	})

	router.Static("/frontend", "/dist")
}
