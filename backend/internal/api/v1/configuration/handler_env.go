package configuration

import (
	"errors"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (handler *Handler) GetEnvConfigHandler(c *gin.Context) {
	loggerModel := handler.createLog(logger.LoggerModel{
		Name:        "GetEnvConfig",
		Description: "Fetching .env configuration",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	})

	if handler.service == nil {
		handler.completeError(loggerModel, errors.New("configuration service is nil"))
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_CONFIGURATION_LOAD_FAILED")})
		return
	}

	envConfig, err := handler.service.GetEnvConfig()
	if err != nil {
		handler.completeError(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_ENV_LOAD_FAILED")})
		return
	}

	handler.completeSuccess(loggerModel)
	c.JSON(http.StatusOK, envConfig)
}

func (handler *Handler) UpdateEnvConfigHandler(c *gin.Context) {
	loggerModel := handler.createLog(logger.LoggerModel{
		Name:        "UpdateEnvConfig",
		Description: "Updating .env configuration",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	})

	if handler.service == nil {
		handler.completeError(loggerModel, errors.New("configuration service is nil"))
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_ENV_UPDATE_FAILED")})
		return
	}

	var request UpdateEnvConfigRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handler.completeError(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	envConfig, err := handler.service.UpdateEnvConfig(request)
	if err != nil {
		handler.completeError(loggerModel, err)
		respondEnvError(c, err)
		return
	}

	handler.completeSuccess(loggerModel)
	c.JSON(http.StatusOK, envConfig)
}

func (handler *Handler) TestEnvDatabaseHandler(c *gin.Context) {
	loggerModel := handler.createLog(logger.LoggerModel{
		Name:        "TestEnvDatabase",
		Description: "Testing candidate database connection",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	})

	if handler.service == nil {
		handler.completeError(loggerModel, errors.New("configuration service is nil"))
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_ENV_LOAD_FAILED")})
		return
	}

	var request TestDatabaseRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handler.completeError(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	if err := handler.service.TestDatabaseConnection(request); err != nil {
		handler.completeError(loggerModel, err)
		c.JSON(http.StatusOK, EnvTestResultDto{Ok: false, Message: i18n.GetMessage("ENV_TEST_DB_FAILED")})
		return
	}

	handler.completeSuccess(loggerModel)
	c.JSON(http.StatusOK, EnvTestResultDto{Ok: true, Message: i18n.GetMessage("ENV_TEST_DB_OK")})
}

func (handler *Handler) TestEnvPathHandler(c *gin.Context) {
	loggerModel := handler.createLog(logger.LoggerModel{
		Name:        "TestEnvPath",
		Description: "Testing candidate filesystem path",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	})

	if handler.service == nil {
		handler.completeError(loggerModel, errors.New("configuration service is nil"))
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_ENV_LOAD_FAILED")})
		return
	}

	var request TestPathRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		handler.completeError(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	if err := handler.service.TestPath(request); err != nil {
		handler.completeError(loggerModel, err)
		c.JSON(http.StatusOK, EnvTestResultDto{Ok: false, Message: i18n.GetMessage("ENV_TEST_PATH_FAILED")})
		return
	}

	handler.completeSuccess(loggerModel)
	c.JSON(http.StatusOK, EnvTestResultDto{Ok: true, Message: i18n.GetMessage("ENV_TEST_PATH_OK")})
}

func respondEnvError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidEnvKey):
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_ENV_INVALID_KEY")})
	case errors.Is(err, ErrInvalidEnvValue):
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_ENV_INVALID_VALUE")})
	case errors.Is(err, ErrEnvConfirmationRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_ENV_CONFIRMATION_REQUIRED")})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_ENV_UPDATE_FAILED")})
	}
}
