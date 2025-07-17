// Package main provides ...
package main

import (
	"path/filepath"

	"github.com/eiblog/eiblog/cmd/eiblog/config"
	"github.com/eiblog/eiblog/cmd/eiblog/handler/admin"
	"github.com/eiblog/eiblog/cmd/eiblog/handler/file"
	"github.com/eiblog/eiblog/cmd/eiblog/handler/page"
	"github.com/eiblog/eiblog/cmd/eiblog/handler/swag"

	"github.com/eiblog/eiblog/pkg/middleware"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.Info("Hi, it's App " + config.Conf.Name)

	endRun := make(chan error, 1)

	runHTTPServer(endRun)
	logrus.Fatal(<-endRun)
}

func runHTTPServer(endRun chan error) {
	if config.Conf.RunMode.IsReleaseMode() {
		gin.SetMode(gin.ReleaseMode)
	}
	e := gin.Default()
	// middleware
	e.Use(middleware.UserMiddleware())
	e.Use(middleware.SessionMiddleware(
		middleware.SessionOpts{
			Name:   "su",
			Secure: config.Conf.RunMode.IsReleaseMode(),
			Secret: []byte("ZGlzvcmUoMTAsICI="),
		}))

	// swag
	swag.RegisterRoutes(e)

	// static files, page
	e.Static("/static", filepath.Join(config.EtcDir, "assets"))
	// custom pages
	e.Static("/page", filepath.Join(config.EtcDir, "page"))

	// static files
	file.RegisterRoutes(e)
	// frontend pages
	page.RegisterRoutes(e)
	// unauthz api
	admin.RegisterRoutes(e)

	// admin router
	group := e.Group("/admin", middleware.AuthFilter)
	{
		page.RegisterRoutesAuthz(group)
		admin.RegisterRoutesAuthz(group)
	}

	// start
	go func() {
		endRun <- e.Run(config.Conf.Listen)
	}()
	logrus.Info("HTTP server running on: " + config.Conf.Listen)
}
