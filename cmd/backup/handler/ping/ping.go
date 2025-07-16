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
// @Summary ping
// @Description ping
// @Tags ping
// @Accept json
// @Produce json
// @Success 200 {string} string "it's ok"
// @Router /ping [get]
func handlePing(c *gin.Context) {
	c.String(http.StatusOK, "it's ok")
}
