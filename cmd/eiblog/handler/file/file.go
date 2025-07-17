package file

import (
	"net/http"
	"path/filepath"

	"github.com/eiblog/eiblog/cmd/eiblog/config"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func RegisterRoutes(e *gin.Engine) {
	e.GET("/rss.html", handleFeed())
	e.GET("/feed", handleFeed())
	e.GET("/opensearch.xml", handleOpensearch())
	e.GET("/sitemap.xml", handleSitemap())
	e.GET("/robots.txt", handleRobots())
	e.GET("/crossdomain.xml", handleCrossDomain())
	e.GET("/favicon.ico", handleFavicon())
}

// handleFeed feed.xml
func handleFeed() gin.HandlerFunc {
	path := filepath.Join(config.EtcDir, "assets", "feed.xml")
	return func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, path)
	}
}

// handleOpensearch opensearch.xml
func handleOpensearch() gin.HandlerFunc {
	path := filepath.Join(config.EtcDir, "assets", "opensearch.xml")
	return func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, path)
	}
}

// handleRobots robotx.txt
func handleRobots() gin.HandlerFunc {
	path := filepath.Join(config.EtcDir, "assets", "robots.txt")
	return func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, path)
	}
}

// handleSitemap sitemap.xml
func handleSitemap() gin.HandlerFunc {
	path := filepath.Join(config.EtcDir, "assets", "sitemap.xml")
	return func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, path)
	}
}

// handleCrossDomain crossdomain.xml
func handleCrossDomain() gin.HandlerFunc {
	path := filepath.Join(config.EtcDir, "assets", "crossdomain.xml")
	return func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, path)
	}
}

// handleFavicon favicon.ico
func handleFavicon() gin.HandlerFunc {
	path := filepath.Join(config.EtcDir, "assets", "favicon.ico")
	return func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, path)
	}
}
