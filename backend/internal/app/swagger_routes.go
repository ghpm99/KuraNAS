package app

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

//go:embed swagger/openapi.json
var openAPISpec []byte

func registerSwaggerRoutes(router *gin.Engine) {
	router.GET("/api-docs/openapi.json", func(context *gin.Context) {
		context.Data(http.StatusOK, "application/json; charset=utf-8", openAPISpec)
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.URL("/api-docs/openapi.json"),
		ginSwagger.DocExpansion("none"),
		ginSwagger.DefaultModelsExpandDepth(-1),
	))
}
