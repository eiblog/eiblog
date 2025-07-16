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

// WalkWorkDir walk work dir
func WalkWorkDir() (string, error) {
	gopath := os.Getenv("GOPATH")
	workDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	// find work dir, try 3 times
	for gopath != workDir && workDir != "/" {
		dir := filepath.Join(workDir, "etc")

		_, err := os.Stat(dir)
		if err == nil {
			break
		}
		if !os.IsNotExist(err) {
			return "", err
		}
		workDir = filepath.Dir(workDir)
	}
	return workDir, nil
}
