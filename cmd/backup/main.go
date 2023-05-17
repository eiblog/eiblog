// Package main provides ...
package main

import (
	"flag"
	"fmt"

	"github.com/eiblog/eiblog/pkg/config"
	"github.com/eiblog/eiblog/pkg/core/backup/ping"
	"github.com/eiblog/eiblog/pkg/core/backup/swag"
	"github.com/eiblog/eiblog/pkg/core/backup/timer"

	"github.com/gin-gonic/gin"
)

var restore bool

func init() {
	flag.BoolVar(&restore, "restore", false, "restore data into mongodb")
}

func main() {
	fmt.Println("Hi, it's App " + config.Conf.BackupApp.Name)
	flag.Parse()

	endRun := make(chan error, 1)

	runCommand(restore, endRun)

	runHTTPServer(endRun)
	fmt.Println(<-endRun)
}

func runCommand(restore bool, endRun chan error) {
	go func() {
		endRun <- timer.Start(restore)
	}()
}

func runHTTPServer(endRun chan error) {
	if !config.Conf.BackupApp.EnableHTTP {
		return
	}

	if config.Conf.RunMode == config.ModeProd {
		gin.SetMode(gin.ReleaseMode)
	}
	e := gin.Default()

	// swag
	swag.RegisterRoutes(e)

	// route
	ping.RegisterRoutes(e)

	// start
	address := fmt.Sprintf(":%d", config.Conf.BackupApp.HTTPPort)
	go func() {
		endRun <- e.Run(address)
	}()
	fmt.Println("HTTP server running on: " + address)
}
