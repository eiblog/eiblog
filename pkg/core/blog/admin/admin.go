// Package admin provides ...
package admin

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/eiblog/eiblog/pkg/cache"
	"github.com/eiblog/eiblog/pkg/config"
	"github.com/eiblog/eiblog/pkg/core/blog"
	"github.com/eiblog/eiblog/pkg/internal"
	"github.com/eiblog/eiblog/pkg/model"
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
	group.POST("/api/account", handleAPIAccount)
	group.POST("/api/blog", handleAPIBlogger)
	group.POST("/api/password", handleAPIPassword)
	group.POST("/api/post-delete", handleAPIPostDelete)
	group.POST("/api/post-add", handleAPIPostCreate)
	group.POST("/api/serie-delete", handleAPISerieDelete)
	group.POST("/api/serie-add", handleAPISerieCreate)
	group.POST("/api/serie-sort", handleAPISerieSort)
	group.POST("/api/draft-delete", handleDraftDelete)
	group.POST("/api/trash-delete", handleAPITrashDelete)
	group.POST("/api/trash-recover", handleAPITrashRecover)
	group.POST("/api/file-upload", handleAPIQiniuUpload)
	group.POST("/api/file-delete", handleAPIQiniuDelete)
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

// handleDraftDelete 删除草稿, 物理删除
func handleDraftDelete(c *gin.Context) {
	for _, v := range c.PostFormArray("mid[]") {
		id, err := strconv.Atoi(v)
		if err != nil || id < 1 {
			responseNotice(c, NoticeNotice, "参数错误", "")
			return
		}
		err = cache.Ei.RemoveArticle(context.Background(), id)
		if err != nil {
			logrus.Error("handleDraftDelete.RemoveArticle: ", err)
			responseNotice(c, NoticeNotice, "删除失败", "")
			return
		}
	}
	responseNotice(c, NoticeSuccess, "删除成功", "")
}

// handleAPIPostDelete 删除文章，移入回收箱
func handleAPIPostDelete(c *gin.Context) {
	var ids []int
	for _, v := range c.PostFormArray("cid[]") {
		id, err := strconv.Atoi(v)
		if err != nil || id < config.Conf.BlogApp.General.StartID {
			responseNotice(c, NoticeNotice, "参数错误", "")
			return
		}
		err = cache.Ei.DelArticle(id)
		if err != nil {
			logrus.Error("handleAPIPostDelete.DelArticle: ", err)

			responseNotice(c, NoticeNotice, err.Error(), "")
			return
		}
		ids = append(ids, id)
	}
	// elasticsearch
	err := internal.ElasticDelIndex(ids)
	if err != nil {
		logrus.Error("handleAPIPostDelete.ElasticDelIndex: ", err)
	}
	// TODO disqus delete
	responseNotice(c, NoticeSuccess, "删除成功", "")
}

// handleAPIPostCreate 创建文章
func handleAPIPostCreate(c *gin.Context) {
	var (
		err error
		do  string
		cid int
	)
	defer func() {
		now := time.Now().In(tools.TimeLocation)
		switch do {
		case "auto": // 自动保存
			if err != nil {
				c.JSON(http.StatusOK, gin.H{"fail": 1, "time": now.Format("15:04:05 PM"), "cid": cid})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": 0, "time": now.Format("15:04:05 PM"), "cid": cid})
		case "save", "publish": // 草稿，发布
			if err != nil {
				responseNotice(c, NoticeNotice, err.Error(), "")
				return
			}
			uri := "/admin/manage-draft"
			if do == "publish" {
				uri = "/admin/manage-posts"
			}
			c.Redirect(http.StatusFound, uri)
		}
	}()

	do = c.PostForm("do") // auto or save or publish
	slug := c.PostForm("slug")
	title := c.PostForm("title")
	text := c.PostForm("text")
	date := parseLocationDate(c.PostForm("date"))
	serie := c.PostForm("serie")
	tag := c.PostForm("tags")
	update := c.PostForm("update")
	if slug == "" || title == "" || text == "" {
		err = errors.New("参数错误")
		return
	}
	var tags []string
	if tag != "" {
		tags = strings.Split(tag, ",")
	}
	serieid, _ := strconv.Atoi(serie)
	article := &model.Article{
		Title:     title,
		Content:   text,
		Slug:      slug,
		IsDraft:   do != "publish",
		Author:    cache.Ei.Account.Username,
		SerieID:   serieid,
		Tags:      tags,
		CreatedAt: date,
	}
	cid, err = strconv.Atoi(c.PostForm("cid"))
	// 新文章
	if err != nil || cid < 1 {
		err = cache.Ei.AddArticle(article)
		if err != nil {
			logrus.Error("handleAPIPostCreate.AddArticle: ", err)
			return
		}
		if !article.IsDraft {
			// 异步执行，快
			go func() {
				// elastic
				internal.ElasticAddIndex(article)
				// rss
				internal.PingFunc(cache.Ei.Blogger.BTitle, slug)
				// disqus
				internal.ThreadCreate(article, cache.Ei.Blogger.BTitle)
			}()
		}
		return
	}
	// 旧文章
	article.ID = cid
	artc, _ := cache.Ei.FindArticleByID(article.ID)
	if artc != nil {
		article.IsDraft = false
		article.Count = artc.Count
		article.UpdatedAt = artc.UpdatedAt
	}
	if update == "true" || update == "1" {
		article.UpdatedAt = time.Now()
	}
	// 数据库更新
	err = cache.Ei.UpdateArticle(context.Background(), article.ID, map[string]interface{}{
		"title":      article.Title,
		"content":    article.Content,
		"serie_id":   article.SerieID,
		"is_draft":   article.IsDraft,
		"tags":       article.Tags,
		"updated_at": article.UpdatedAt,
		"created_at": article.CreatedAt,
	})
	if err != nil {
		logrus.Error("handleAPIPostCreate.UpdateArticle: ", err)
		return
	}
	if !article.IsDraft {
		cache.Ei.RepArticle(artc, article)
		// 异步执行，快
		go func() {
			// elastic
			internal.ElasticAddIndex(article)
			// rss
			internal.PingFunc(cache.Ei.Blogger.BTitle, slug)
			// disqus
			if artc == nil {
				internal.ThreadCreate(article, cache.Ei.Blogger.BTitle)
			}
		}()
	}
}

// handleAPISerieDelete 只能逐一删除，专题下不能有文章
func handleAPISerieDelete(c *gin.Context) {
	for _, v := range c.PostFormArray("mid[]") {
		id, err := strconv.Atoi(v)
		if err != nil || id < 1 {
			responseNotice(c, NoticeNotice, err.Error(), "")
			return
		}
		err = cache.Ei.DelSerie(id)
		if err != nil {
			responseNotice(c, NoticeNotice, err.Error(), "")
			return
		}
	}
	responseNotice(c, NoticeSuccess, "删除成功", "")
}

// handleAPISerieSort 专题排序
func handleAPISerieSort(c *gin.Context) {
	v := c.PostFormArray("mid[]")
	logrus.Debug(v)
}

// handleAPISerieCreate 添加专题，如果专题有提交 mid 即更新专题
func handleAPISerieCreate(c *gin.Context) {
	name := c.PostForm("name")
	slug := c.PostForm("slug")
	desc := c.PostForm("description")
	if name == "" || slug == "" || desc == "" {
		responseNotice(c, NoticeNotice, "参数错误", "")
		return
	}
	mid, err := strconv.Atoi(c.PostForm("mid"))
	if err == nil && mid > 0 {
		var serie *model.Serie
		for _, v := range cache.Ei.Series {
			if v.ID == mid {
				serie = v
				break
			}
		}
		if serie == nil {
			responseNotice(c, NoticeNotice, "专题不存在", "")
			return
		}
		err = cache.Ei.UpdateSerie(context.Background(), mid, map[string]interface{}{
			"slug": slug,
			"name": name,
			"desc": desc,
		})
		if err != nil {
			logrus.Error("handleAPISerieCreate.UpdateSerie: ", err)
			responseNotice(c, NoticeNotice, err.Error(), "")
			return
		}
		serie.Slug = slug
		serie.Name = name
		serie.Desc = desc
	} else {
		err = cache.Ei.AddSerie(&model.Serie{
			Slug:      slug,
			Name:      name,
			Desc:      desc,
			CreatedAt: time.Now(),
		})
		if err != nil {
			logrus.Error("handleAPISerieCreate.InsertSerie: ", err)
			responseNotice(c, NoticeNotice, err.Error(), "")
			return
		}
	}
	responseNotice(c, NoticeSuccess, "操作成功", "")
}

// handleAPITrashDelete 删除回收箱, 物理删除
func handleAPITrashDelete(c *gin.Context) {
	for _, v := range c.PostFormArray("mid[]") {
		id, err := strconv.Atoi(v)
		if err != nil || id < 1 {
			responseNotice(c, NoticeNotice, "参数错误", "")
			return
		}
		err = cache.Ei.RemoveArticle(context.Background(), id)
		if err != nil {
			responseNotice(c, NoticeNotice, err.Error(), "")
			return
		}
	}
	responseNotice(c, NoticeSuccess, "删除成功", "")
}

// handleAPITrashRecover 恢复到草稿
func handleAPITrashRecover(c *gin.Context) {
	for _, v := range c.PostFormArray("mid[]") {
		id, err := strconv.Atoi(v)
		if err != nil || id < 1 {
			responseNotice(c, NoticeNotice, "参数错误", "")
			return

		}
		err = cache.Ei.UpdateArticle(context.Background(), id, map[string]interface{}{
			"deleted_at": time.Time{},
			"is_draft":   true,
		})
		if err != nil {
			responseNotice(c, NoticeNotice, err.Error(), "")
			return
		}
	}
	responseNotice(c, NoticeSuccess, "恢复成功", "")
}

// handleAPIQiniuUpload 上传文件
func handleAPIQiniuUpload(c *gin.Context) {
	type Size interface {
		Size() int64
	}
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		logrus.Error("handleAPIQiniuUpload.FormFile: ", err)
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	s, ok := file.(Size)
	if !ok {
		logrus.Error("assert failed")
		c.String(http.StatusBadRequest, "false")
		return
	}
	filename := strings.ToLower(header.Filename)
	url, err := internal.QiniuUpload(filename, s.Size(), file)
	if err != nil {
		logrus.Error("handleAPIQiniuUpload.QiniuUpload: ", err)
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	typ := header.Header.Get("Content-Type")
	c.JSON(http.StatusOK, gin.H{
		"title":   filename,
		"isImage": typ[:5] == "image",
		"url":     url,
		"bytes":   fmt.Sprintf("%dkb", s.Size()/1000),
	})
}

// handleAPIQiniuDelete 删除文件
func handleAPIQiniuDelete(c *gin.Context) {
	defer c.String(http.StatusOK, "删掉了吗？鬼知道。。。")

	name := c.PostForm("title")
	if name == "" {
		logrus.Error("handleAPIQiniuDelete.PostForm: 参数错误")
		return
	}
	err := internal.QiniuDelete(name)
	if err != nil {
		logrus.Error("handleAPIQiniuDelete.QiniuDelete: ", err)
	}
}

// parseLocationDate 解析日期
func parseLocationDate(date string) time.Time {
	t, err := time.ParseInLocation("2006-01-02 15:04", date, tools.TimeLocation)
	if err == nil {
		return t.UTC()
	}
	return time.Now()
}

func responseNotice(c *gin.Context, typ, content, hl string) {
	if hl != "" {
		c.SetCookie("notice_highlight", hl, 86400, "/", "", true, false)
	}
	c.SetCookie("notice_type", typ, 86400, "/", "", true, false)
	c.SetCookie("notice", fmt.Sprintf("[\"%s\"]", content), 86400, "/", "", true, false)
	c.Redirect(http.StatusFound, c.Request.Referer())
}
