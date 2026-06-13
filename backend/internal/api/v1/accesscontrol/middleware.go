package accesscontrol

import (
	"net/http"
	"net/netip"

	"nas-go/api/pkg/i18n"

	"github.com/gin-gonic/gin"
)

// NewMiddleware blocks every request whose origin IP is neither loopback nor
// covered by an enabled whitelist entry. It must be registered on the engine
// before any route (API, assets, SPA, Swagger).
//
// The decision uses c.RemoteIP() — the IP of the actual TCP connection — and
// never proxy headers, so a forged X-Forwarded-For/X-Real-IP cannot bypass the
// whitelist (the app also calls SetTrustedProxies(nil) at boot). Loopback
// always passes: the owner can never lock themselves out on the server itself.
func NewMiddleware(service ServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		remoteIP := c.RemoteIP()
		addr, err := netip.ParseAddr(remoteIP)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": i18n.Translate("ERROR_IP_NOT_ALLOWED", remoteIP),
				"ip":    remoteIP,
			})
			return
		}

		// Unmap normalizes IPv4-mapped IPv6; WithZone("") strips the scope id
		// (e.g. fe80::1%eth0) that a link-local connection carries, because
		// netip.Prefix.Contains never matches a zoned address ("Prefixes strip
		// zones") and would otherwise block every link-local device.
		addr = addr.Unmap().WithZone("")
		if addr.IsLoopback() {
			c.Next()
			return
		}

		if service != nil && service.IsAllowed(addr) {
			c.Next()
			return
		}

		// The requester IP goes in the body to make registering the device a
		// one-click action — it tells the client nothing it does not know.
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": i18n.Translate("ERROR_IP_NOT_ALLOWED", addr.String()),
			"ip":    addr.String(),
		})
	}
}
