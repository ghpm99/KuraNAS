package app

import (
	"database/sql"
	"nas-go/api/internal/api/v1/files"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, database *sql.DB) {

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

	routesV1 := router.Group("/api/v1")
	files.RegisterFilesRoutes(routesV1, database)
	RegisterReactRoutes(router)
}

func RegisterReactRoutes(router *gin.Engine) {
	router.Static("/assets", "./dist/assets")

	router.NoRoute(func(c *gin.Context) {
		c.File("./dist/index.html")
	})

	router.Static("/frontend", "/dist")
}
