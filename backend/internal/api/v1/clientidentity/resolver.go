package clientidentity

import (
	"regexp"

	"github.com/gin-gonic/gin"
)

const (
	Header     = "X-KuraNAS-Client-Id"
	queryParam = "client_id"
)

var clientIDPattern = regexp.MustCompile(`^[A-Za-z0-9-]{8,64}$`)

// Resolve returns the per-device client id declared by the request (header, then
// query parameter), falling back to the caller IP when none is declared. The
// boolean is false when a declared id is malformed.
func Resolve(c *gin.Context) (string, bool) {
	declaredClientID := c.GetHeader(Header)
	if declaredClientID == "" {
		declaredClientID = c.Query(queryParam)
	}
	if declaredClientID == "" {
		return c.ClientIP(), true
	}
	return declaredClientID, clientIDPattern.MatchString(declaredClientID)
}
