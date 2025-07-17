// Package config provides ...
package config

import (
	"os"
	"path/filepath"
)

// RunMode 列表
const (
	RunModeLocal RunMode = "local" // 本地环境
	RunModeDev   RunMode = "dev"   // 开发环境
	RunModeProd  RunMode = "pro"   // 生产环境
)

// RunMode 运行模式
type RunMode string

// IsReleaseMode 是否
func (mode RunMode) IsReleaseMode() bool {
	return mode == RunModeProd
}

// IsRunMode 是否是runmode
func (mode RunMode) IsRunMode() bool {
	return mode == RunModeDev || mode == RunModeProd || mode == RunModeLocal
}

// WorkEtcPath walk etc dir
func WorkEtcPath() (string, error) {
	gopath := os.Getenv("GOPATH")
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	// find etc path, try 3 times
	var etc string
	for gopath != wd && wd != "/" {
		etc = filepath.Join(wd, "etc")

		_, err := os.Stat(etc)
		if err == nil {
			break
		}
		if !os.IsNotExist(err) {
			return "", err
		}
		wd = filepath.Dir(wd)
	}
	return etc, nil
}
