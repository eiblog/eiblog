package page

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes register routes
func RegisterRoutes(e *gin.Engine) {
	e.NoRoute(handleNotFound)

	e.GET("/", handleHomePage)
	e.GET("/post/:slug", handleArticlePage)
	e.GET("/series.html", handleSeriesPage)
	e.GET("/archives.html", handleArchivePage)
	e.GET("/search.html", handleSearchPage)
	e.GET("/disqus/post-:slug", handleDisqusList)
	e.GET("/disqus/form/post-:slug", handleDisqusPage)
	e.POST("/disqus/create", handleDisqusCreate)
	e.GET("/beacon.html", handleBeaconPage)

	// login page
	e.GET("/admin/login", handleLoginPage)
}

// RegisterRoutesAuthz register admin
func RegisterRoutesAuthz(group gin.IRoutes) {
	// console
	group.GET("/profile", handleAdminProfile)
	// write
	group.GET("/write-post", handleAdminPost)
	group.GET("/draft-delete", handleDraftDelete)
	// manage
	group.GET("/manage-posts", handleAdminPosts)
	group.GET("/manage-series", handleAdminSeries)
	group.GET("/add-serie", handleAdminSerie)
	group.GET("/manage-tags", handleAdminTags)
	group.GET("/manage-draft", handleAdminDraft)
	group.GET("/manage-trash", handleAdminTrash)
	group.GET("/options-general", handleAdminGeneral)
	group.GET("/options-discussion", handleAdminDiscussion)
}
