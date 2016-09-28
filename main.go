// Package main provides ...
package main

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/EiBlog/utils/logd"
)

func main() {
	// set log print level
	logd.SetLevel(logd.Ldebug)
	// pprof
	go func() {
		http.ListenAndServe(":6060", nil)
	}()
	Run()
}
