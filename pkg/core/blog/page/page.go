// Package page provides ...
package page

import (
	"path/filepath"
	"text/template"

	"github.com/eiblog/eiblog/pkg/config"
	"github.com/eiblog/eiblog/tools"

	"github.com/eiblog/utils/tmpl"
	"github.com/gin-gonic/gin"
)

// htmlTmpl html template cache
var htmlTmpl *template.Template

func init() {
	htmlTmpl = template.New("eiblog").Funcs(tmpl.TplFuncMap)
	root := filepath.Join(config.WorkDir, "website")
	files := tools.ReadDirFiles(root, func(name string) bool {
		if name == ".DS_Store" {
			return true
		}
		return false
	})
	_, err := htmlTmpl.ParseFiles(files...)
	if err != nil {
		panic(err)
	}
}

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

}
