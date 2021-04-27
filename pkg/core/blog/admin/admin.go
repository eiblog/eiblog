// Package admin provides ...
package admin

import (
	"context"
	"net/http"
	"time"

	"github.com/eiblog/eiblog/pkg/cache"
	"github.com/eiblog/eiblog/pkg/core/blog"
	"github.com/eiblog/eiblog/tools"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RegisterRoutes register routes
func RegisterRoutes(e *gin.Engine) {
	e.POST("/admin/login", handleAcctLogin)
}

// RegisterRoutesAuthz register routes
func RegisterRoutesAuthz(group gin.IRoutes) {
}

// handleAcctLogin 登录接口
func handleAcctLogin(c *gin.Context) {
	user := c.PostForm("user")
	pwd := c.PostForm("password")
	// code := c.PostForm("code") // 二次验证
	if user == "" || pwd == "" {
		logrus.Warnf("参数错误: %s %s", user, pwd)
		c.Redirect(http.StatusFound, "/admin/login")
		return
	}
	if cache.Ei.Account.Username != user ||
		cache.Ei.Account.Password != tools.EncryptPasswd(user, pwd) {
		logrus.Warnf("账号或密码错误 %s, %s", user, pwd)
		c.Redirect(http.StatusFound, "/admin/login")
		return
	}
	// 登录成功
	blog.SetLogin(c, user)

	cache.Ei.Account.LoginIP = c.ClientIP()
	cache.Ei.Account.LoginAt = time.Now()
	cache.Ei.UpdateAccount(context.Background(), user, map[string]interface{}{
		"login_ip": cache.Ei.Account.LoginIP,
		"login_at": cache.Ei.Account.LoginAt,
	})
	c.Redirect(http.StatusFound, "/admin/profile")
}
