// Package main provides ...
package main

import (
	"flag"

	"github.com/eiblog/eiblog/cmd/backup/config"
	"github.com/eiblog/eiblog/cmd/backup/handler/ping"
	"github.com/eiblog/eiblog/cmd/backup/handler/swag"
	"github.com/eiblog/eiblog/cmd/backup/handler/timer"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// @title backup API
// @version 1.0
// @description This is a backup server.

var restore bool

func init() {
	flag.BoolVar(&restore, "restore", false, "restore data into mongodb")
}

func main() {
	logrus.Info("Hi, it's App " + config.Conf.Name)

	flag.Parse()
	endRun := make(chan error, 1)

	runCommand(restore, endRun)

	runHTTPServer(endRun)
	logrus.Fatal(<-endRun)
}

func runCommand(restore bool, endRun chan error) {
	go func() {
		endRun <- timer.Start(restore)
	}()
}

func runHTTPServer(endRun chan error) {
	if config.Conf.RunMode.IsReleaseMode() {
		gin.SetMode(gin.ReleaseMode)
	}
	e := gin.Default()

	// swag
	swag.RegisterRoutes(e)

	// route
	ping.RegisterRoutes(e)

	// start
	go func() {
		endRun <- e.Run(config.Conf.Listen)
	}()
	logrus.Info("HTTP server running on: " + config.Conf.Listen)
}
