// Package main provides ...
package main

import (
	"fmt"
	"path/filepath"

	"github.com/eiblog/eiblog/pkg/config"
	"github.com/eiblog/eiblog/pkg/core/eiblog"
	"github.com/eiblog/eiblog/pkg/core/eiblog/admin"
	"github.com/eiblog/eiblog/pkg/core/eiblog/file"
	"github.com/eiblog/eiblog/pkg/core/eiblog/page"
	"github.com/eiblog/eiblog/pkg/core/eiblog/swag"
	"github.com/eiblog/eiblog/pkg/mid"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("Hi, it's App " + config.Conf.EiBlogApp.Name)

	endRun := make(chan bool, 1)

	runHTTPServer(endRun)
	<-endRun
}

func runHTTPServer(endRun chan bool) {
	if !config.Conf.EiBlogApp.EnableHTTP {
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
	group := e.Group("/admin", eiblog.AuthFilter)
	{
		page.RegisterRoutesAuthz(group)
		admin.RegisterRoutesAuthz(group)
	}

	// start
	address := fmt.Sprintf(":%d", config.Conf.EiBlogApp.HTTPPort)
	go e.Run(address)
	fmt.Println("HTTP server running on: " + address)
}
