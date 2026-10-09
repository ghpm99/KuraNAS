package music

import (
	"regexp"

	"github.com/gin-gonic/gin"
)

const (
	PlayerClientIDHeader     = "X-KuraNAS-Client-Id"
	playerClientIDQueryParam = "client_id"
)

var playerClientIDPattern = regexp.MustCompile(`^[A-Za-z0-9-]{8,64}$`)

func resolvePlayerClientID(c *gin.Context) (string, bool) {
	declaredClientID := c.GetHeader(PlayerClientIDHeader)
	if declaredClientID == "" {
		declaredClientID = c.Query(playerClientIDQueryParam)
	}
	if declaredClientID == "" {
		return c.ClientIP(), true
	}
	return declaredClientID, playerClientIDPattern.MatchString(declaredClientID)
}
