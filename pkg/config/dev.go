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
			dir, _ := path.Split(wd)
			wd = path.Clean(dir)
			continue
		}
		return wd
	}
	return ""
}
