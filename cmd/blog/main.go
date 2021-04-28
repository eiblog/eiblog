// Package main provides ...
package main

import (
	"fmt"
	"path/filepath"

	"github.com/eiblog/eiblog/pkg/config"
	"github.com/eiblog/eiblog/pkg/core/blog"
	"github.com/eiblog/eiblog/pkg/core/blog/admin"
	"github.com/eiblog/eiblog/pkg/core/blog/file"
	"github.com/eiblog/eiblog/pkg/core/blog/page"
	"github.com/eiblog/eiblog/pkg/core/blog/swag"
	"github.com/eiblog/eiblog/pkg/mid"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("Hi, it's App Demo")

	endRun := make(chan bool, 1)

	runHTTPServer(endRun)
	<-endRun
}

func runHTTPServer(endRun chan bool) {
	if !config.Conf.BlogApp.EnableHTTP {
		return
	}

	if config.Conf.RunMode == config.ModeProd {
		gin.SetMode(gin.ReleaseMode)
	}
	e := gin.Default()
	// middleware
	e.Use(mid.UserMiddleware())
	e.Use(mid.SessionMiddleware(mid.SessionOpts{
		Name:   "su",
		Secure: config.Conf.RunMode == config.ModeProd,
		Secret: []byte("ZGlzvcmUoMTAsICI="),
	}))

	// swag
	swag.RegisterRoutes(e)

	// static files, page
	root := filepath.Join(config.WorkDir, "assets")
	e.Static("/static", root)

	// static files
	file.RegisterRoutes(e)
	// frontend pages
	page.RegisterRoutes(e)
	// unauthz api
	admin.RegisterRoutes(e)

	// admin router
	group := e.Group("/admin", blog.AuthFilter)
	{
		page.RegisterRoutesAuthz(group)
		admin.RegisterRoutesAuthz(group)
	}

	// start
	address := fmt.Sprintf(":%d", config.Conf.BlogApp.HTTPPort)
	go e.Run(address)
	fmt.Println("HTTP server running on: " + address)
}
