package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// NoCacheMiddleware disables caching for static files in debug mode
func NoCacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if gin.Mode() == gin.DebugMode && strings.HasPrefix(c.Request.URL.Path, "/static/") {
			c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		}
		c.Next()
	}
}
