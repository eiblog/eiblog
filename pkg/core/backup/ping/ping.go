// Package ping provides ...
package ping

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func RegisterRoutes(group gin.IRoutes) {
	group.GET("/ping", handlePing)
}

// handlePing ping
func handlePing(c *gin.Context) {
	c.String(http.StatusOK, "it's ok")
}
