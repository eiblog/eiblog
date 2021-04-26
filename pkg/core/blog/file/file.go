// Package file provides ...
package file

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func RegisterRoutes(e *gin.Engine) {
	e.GET("/rss.html", handleFeed)
	e.GET("/feed", handleFeed)
	e.GET("/opensearch.xml", handleOpensearch)
	e.GET("/sitemap.xml", handleSitemap)
	e.GET("/robots.txt", handleRobots)
	e.GET("/crossdomain.xml", handleCrossDomain)
	e.GET("/favicon.ico", handleFavicon)
}

// handleFeed feed.xml
func handleFeed(c *gin.Context) {
	http.ServeFile(c.Writer, c.Request, "assets/feed.xml")
}

// handleOpensearch opensearch.xml
func handleOpensearch(c *gin.Context) {
	http.ServeFile(c.Writer, c.Request, "assets/opensearch.xml")
}

// handleRobots robotx.txt
func handleRobots(c *gin.Context) {
	http.ServeFile(c.Writer, c.Request, "assets/robots.txt")
}

// handleSitemap sitemap.xml
func handleSitemap(c *gin.Context) {
	http.ServeFile(c.Writer, c.Request, "assets/sitemap.xml")
}

// handleCrossDomain crossdomain.xml
func handleCrossDomain(c *gin.Context) {
	http.ServeFile(c.Writer, c.Request, "assets/crossdomain.xml")
}

// handleFavicon favicon.ico
func handleFavicon(c *gin.Context) {
	http.ServeFile(c.Writer, c.Request, "assets/favicon.ico")
}
