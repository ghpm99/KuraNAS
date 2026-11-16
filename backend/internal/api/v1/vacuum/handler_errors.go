package vacuum

import (
	"errors"
	"net/http"

	"nas-go/api/pkg/applog"
	"nas-go/api/pkg/i18n"

	"github.com/gin-gonic/gin"
)

type errorResponseRule struct {
	cause      error
	statusCode int
	messageKey string
}

var errorResponseRules = []errorResponseRule{
	{ErrNotConfigured, http.StatusConflict, "ERROR_VACUUM_NOT_CONFIGURED"},
	{ErrDeviceUnreachable, http.StatusBadGateway, "ERROR_VACUUM_UNREACHABLE"},
	{ErrLocalKeyRejected, http.StatusBadGateway, "ERROR_VACUUM_LOCAL_KEY_REJECTED"},
	{ErrDiscoveryNotFound, http.StatusNotFound, "ERROR_VACUUM_DISCOVERY_NOT_FOUND"},
	{ErrInvalidCommand, http.StatusBadRequest, "ERROR_VACUUM_INVALID_COMMAND"},
	{ErrInvalidDirection, http.StatusBadRequest, "ERROR_VACUUM_INVALID_DIRECTION"},
	{ErrInvalidPart, http.StatusBadRequest, "ERROR_VACUUM_INVALID_PART"},
	{ErrInvalidPreference, http.StatusBadRequest, "ERROR_VACUUM_INVALID_PREFERENCE"},
	{ErrInvalidAddress, http.StatusBadRequest, "ERROR_VACUUM_INVALID_ADDRESS"},
	{ErrInvalidDeviceID, http.StatusBadRequest, "ERROR_VACUUM_INVALID_DEVICE_ID"},
	{ErrInvalidLocalKey, http.StatusBadRequest, "ERROR_VACUUM_INVALID_LOCAL_KEY"},
	{ErrLocalKeyRequired, http.StatusBadRequest, "ERROR_VACUUM_LOCAL_KEY_REQUIRED"},
	{ErrInvalidRequest, http.StatusBadRequest, "ERROR_VACUUM_INVALID_REQUEST"},
}

func respondWithError(c *gin.Context, failure error) {
	for _, rule := range errorResponseRules {
		if errors.Is(failure, rule.cause) {
			logExpectedFailure(c, rule, failure)
			c.JSON(rule.statusCode, gin.H{"error": i18n.GetMessage(rule.messageKey)})
			return
		}
	}

	applog.ErrorWithStack("vacuum: request failed", failure, "path", c.FullPath(), "ip", c.ClientIP())
	c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_VACUUM_FAILED")})
}

func logExpectedFailure(c *gin.Context, rule errorResponseRule, failure error) {
	if rule.statusCode == http.StatusBadRequest {
		return
	}
	applog.Warn("vacuum: request failed", "path", c.FullPath(), "status", rule.statusCode, "error", failure.Error())
}

func respondWithInvalidBody(c *gin.Context) {
	respondWithError(c, ErrInvalidRequest)
}
