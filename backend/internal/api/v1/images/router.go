package images

import (
	"database/sql"

	"github.com/gin-gonic/gin"
)

func RegisterImagesRoutes(router *gin.RouterGroup, database *sql.DB) {

	repository := NewRepository(database)
	service := NewService(repository)
	handler := NewHandler(service)

	images := router.Group("/images")

	images.GET("/", handler.GetAllImagesHandler)

}
