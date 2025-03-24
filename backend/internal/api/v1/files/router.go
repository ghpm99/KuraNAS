package files

import (
	"database/sql"

	"github.com/gin-gonic/gin"
)

func RegisterFilesRoutes(router *gin.RouterGroup, database *sql.DB) {

	repository := NewRepository(database)
	service := NewService(repository)
	handler := NewHandler(service)

	files := router.Group("/files")

	files.GET("/", handler.GetFilesHandler)

}
