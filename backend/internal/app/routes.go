package app

import (
	"database/sql"
	"nas-go/api/internal/api/v1/images"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, database *sql.DB) {
	routesV1 := router.Group("/v1")
	images.RegisterImagesRoutes(routesV1, database)
}
