// +build !prod

// Package config provides ...
package config

import (
	"os"
	"path"
	"path/filepath"
)

// workDir recognize workspace dir
var workDir = func() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	for wd != "" {
		name := filepath.Join(wd, "conf")
		_, err := os.Stat(name)
		if err != nil {
			wd, _ = path.Split(wd)
			continue
		}
		return wd
	}
	return ""
}
