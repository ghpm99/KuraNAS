package app

import "github.com/gin-gonic/gin"

func SetUpRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.UseRawPath = true
	router.UnescapePathValues = true
	return router
}
