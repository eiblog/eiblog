// Package eiblog provides ...
package eiblog

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// @title APP Demo API
// @version 1.0
// @description This is a sample server celler server.

// @BasePath /api

// LogStatus log status
type LogStatus int

// user log status
var (
	LogStatusOut LogStatus = 0
	LogStatusTFA LogStatus = 1
	LogStatusIn  LogStatus = 2
)

// AuthFilter auth filter
func AuthFilter(c *gin.Context) {
	if !IsLogined(c) {
		c.Abort()
		c.Status(http.StatusUnauthorized)
		return
	}

	c.Next()
}

// SetLogStatus login user
func SetLogStatus(c *gin.Context, uid string, status LogStatus) {
	session := sessions.Default(c)
	session.Set("uid", uid)
	session.Set("status", int(status))
	session.Save()
}

// SetLogout logout user
func SetLogout(c *gin.Context) {
	session := sessions.Default(c)
	session.Set("status", int(LogStatusOut))
	session.Save()
}

// IsLogined account logined
func IsLogined(c *gin.Context) bool {
	status := GetLogStatus(c)
	if status < 0 {
		return false
	}
	return status == LogStatusIn
}

// GetUserID get logined account uuid
func GetUserID(c *gin.Context) string {
	session := sessions.Default(c)
	uid := session.Get("uid")
	if uid == nil {
		return ""
	}
	return uid.(string)
}

// GetLogStatus get account log status
func GetLogStatus(c *gin.Context) LogStatus {
	session := sessions.Default(c)
	status := session.Get("status")
	if status == nil {
		return -1
	}
	return LogStatus(status.(int))
}
