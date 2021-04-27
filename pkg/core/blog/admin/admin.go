// Package admin provides ...
package admin

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/eiblog/eiblog/pkg/cache"
	"github.com/eiblog/eiblog/pkg/core/blog"
	"github.com/eiblog/eiblog/tools"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// 通知cookie
const (
	NoticeSuccess = "success"
	NoticeNotice  = "notice"
	NoticeError   = "error"
)

// RegisterRoutes register routes
func RegisterRoutes(e *gin.Engine) {
	e.POST("/admin/login", handleAcctLogin)
}

// RegisterRoutesAuthz register routes
func RegisterRoutesAuthz(group gin.IRoutes) {
	group.GET("/draft-delete", handleDraftDelete)

	group.POST("/api/account", handleAPIAccount)
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

// handleDraftDelete 删除草稿
func handleDraftDelete(c *gin.Context) {
	id, err := strconv.Atoi(c.Query("cid"))
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	err = cache.Ei.RemoveArticle(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "删除错误"})
		return
	}
	c.Redirect(http.StatusFound, "/admin/write-post")
}

// handleAPIAccount 更新账户信息
func handleAPIAccount(c *gin.Context) {
	e := c.PostForm("email")
	pn := c.PostForm("phoneNumber")
	ad := c.PostForm("address")
	if (e != "" && !tools.ValidateEmail(e)) || (pn != "" &&
		!tools.ValidatePhoneNo(pn)) {
		responseNotice(c, NoticeNotice, "参数错误", "")
		return
	}

	err := cache.Ei.UpdateAccount(context.Background(), cache.Ei.Account.Username,
		map[string]interface{}{
			"email":   e,
			"phone_n": pn,
			"address": ad,
		})
	if err != nil {
		logrus.Error("handleAPIAccount.UpdateAccount: ", err)
		responseNotice(c, NoticeNotice, err.Error(), "")
		return
	}
	cache.Ei.Account.Email = e
	cache.Ei.Account.PhoneN = pn
	cache.Ei.Account.Address = ad
	responseNotice(c, NoticeSuccess, "更新成功", "")
}

// handleAPIBlogger 更新博客信息
func handleAPIBlogger(c *gin.Context) {
	bn := c.PostForm("blogName")
	bt := c.PostForm("bTitle")
	ba := c.PostForm("beiAn")
	st := c.PostForm("subTitle")
	ss := c.PostForm("seriessay")
	as := c.PostForm("archivessay")
	if bn == "" || bt == "" {
		responseNotice(c, NoticeNotice, "参数错误", "")
		return
	}

	err := cache.Ei.UpdateBlogger(context.Background(), map[string]interface{}{
		"blog_name":    bn,
		"b_title":      bt,
		"sub_title":    st,
		"series_say":   ss,
		"archives_say": as,
	})
	if err != nil {
		logrus.Error("handleAPIBlogger.UpdateBlogger: ", err)
		responseNotice(c, NoticeNotice, err.Error(), "")
		return
	}
	cache.Ei.Blogger.BlogName = bn
	cache.Ei.Blogger.BTitle = bt
	cache.Ei.Blogger.BeiAn = ba
	cache.Ei.Blogger.SubTitle = st
	cache.Ei.Blogger.SeriesSay = ss
	cache.Ei.Blogger.ArchivesSay = as
	cache.PagesCh <- cache.PageSeries
	cache.PagesCh <- cache.PageArchive
	responseNotice(c, NoticeSuccess, "更新成功", "")
}

// handleAPIPassword 更新密码
func handleAPIPassword(c *gin.Context) {
	od := c.PostForm("old")
	nw := c.PostForm("new")
	cf := c.PostForm("confirm")
	if nw != cf {
		responseNotice(c, NoticeNotice, "两次密码输入不一致", "")
		return
	}
	if !tools.ValidatePassword(nw) {
		responseNotice(c, NoticeNotice, "密码格式错误", "")
		return
	}
	if cache.Ei.Account.Password != tools.EncryptPasswd(cache.Ei.Account.Username, od) {
		responseNotice(c, NoticeNotice, "原始密码不正确", "")
		return
	}
	newPwd := tools.EncryptPasswd(cache.Ei.Account.Username, nw)

	err := cache.Ei.UpdateAccount(context.Background(), cache.Ei.Account.Username,
		map[string]interface{}{
			"password": newPwd,
		})
	if err != nil {
		logrus.Error("handleAPIPassword.UpdateAccount: ", err)
		responseNotice(c, NoticeNotice, err.Error(), "")
		return
	}
	cache.Ei.Account.Password = newPwd
	responseNotice(c, NoticeSuccess, "更新成功", "")
}

// handleAPIPostDelete 删除文章，移入回收箱
func handleAPIPostDelete(c *gin.Context) {
	// var ids []int32
	// for _, v := range c.PostFormArray("cid[]") {
	// 	i, err := strconv.Atoi(v)
	// 	if err != nil || int32(i) < config.Conf.BlogApp.General.StartID {
	// 		responseNotice(c, NoticeNotice, "参数错误", "")
	// 		return
	// 	}
	// 	ids = append(ids, int32(i))
	// }
	// err := DelArticles(ids...)
	// if err != nil {
	// 	logd.Error(err)
	// 	responseNotice(c, NOTICE_NOTICE, err.Error(), "")
	// 	return
	// }
	//
	// // elasticsearch
	// err = ElasticDelIndex(ids)
	// if err != nil {
	// 	logrus.Error("handleAPIPostDelete.")
	// }
	// // TODO disqus delete
	// responseNotice(c, NoticeSuccess, "删除成功", "")
}

func responseNotice(c *gin.Context, typ, content, hl string) {
	if hl != "" {
		c.SetCookie("notice_highlight", hl, 86400, "/", "", true, false)
	}
	c.SetCookie("notice_type", typ, 86400, "/", "", true, false)
	c.SetCookie("notice", fmt.Sprintf("[\"%s\"]", content), 86400, "/", "", true, false)
	c.Redirect(http.StatusFound, c.Request.Referer())
}
