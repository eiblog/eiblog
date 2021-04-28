// Package page provides ...
package page

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	htemplate "html/template"
	"net/http"
	"strconv"

	"github.com/eiblog/eiblog/pkg/cache"
	"github.com/eiblog/eiblog/pkg/cache/store"
	"github.com/eiblog/eiblog/pkg/config"
	"github.com/eiblog/eiblog/pkg/core/blog"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// baseBEParams 基础参数
func baseBEParams(c *gin.Context) gin.H {
	return gin.H{
		"Author": cache.Ei.Account.Username,
		"Qiniu":  config.Conf.BlogApp.Qiniu,
	}
}

// handleLoginPage 登录页面
func handleLoginPage(c *gin.Context) {
	logout := c.Query("logout")
	if logout == "true" {
		blog.SetLogout(c)
	} else if blog.IsLogined(c) {
		c.Redirect(http.StatusFound, "/admin/profile")
		return
	}
	params := gin.H{"BTitle": cache.Ei.Blogger.BTitle}
	renderHTMLAdminLayout(c, "login.html", params)
}

// handleAdminProfile 个人配置
func handleAdminProfile(c *gin.Context) {
	params := baseBEParams(c)
	params["Title"] = "个人配置 | " + cache.Ei.Blogger.BTitle
	params["Path"] = c.Request.URL.Path
	params["Console"] = true
	params["Ei"] = cache.Ei
	renderHTMLAdminLayout(c, "admin-profile", params)
}

type T struct {
	ID   string `json:"id"`
	Tags string `json:"tags"`
}

// handleAdminPost 写文章页
func handleAdminPost(c *gin.Context) {
	params := baseBEParams(c)
	id, err := strconv.Atoi(c.Query("cid"))
	if err == nil && id > 0 {
		article, _ := cache.Ei.LoadArticle(context.Background(), id)
		if article != nil {
			params["Title"] = "编辑文章 | " + cache.Ei.Blogger.BTitle
			params["Edit"] = article
		}
	}
	if params["Title"] == nil {
		params["Title"] = "撰写文章 | " + cache.Ei.Blogger.BTitle
	}
	params["Path"] = c.Request.URL.Path
	params["Domain"] = config.Conf.BlogApp.Host
	params["Series"] = cache.Ei.Series
	var tags []T
	for tag := range cache.Ei.TagArticles {
		tags = append(tags, T{tag, tag})
	}
	str, _ := json.Marshal(tags)
	params["Tags"] = string(str)
	renderHTMLAdminLayout(c, "admin-post", params)
}

// handleAdminPosts 文章管理页
func handleAdminPosts(c *gin.Context) {
	kw := c.Query("keywords")
	tmp := c.Query("serie")
	se, err := strconv.Atoi(tmp)
	if err != nil || se < 1 {
		se = 0
	}
	pg, err := strconv.Atoi(c.Query("page"))
	if err != nil || pg < 1 {
		pg = 1
	}
	vals := c.Request.URL.Query()

	params := baseBEParams(c)
	params["Title"] = "文章管理 | " + cache.Ei.Blogger.BTitle
	params["Manage"] = true
	params["Path"] = c.Request.URL.Path
	params["Series"] = cache.Ei.Series
	params["Serie"] = se
	params["KW"] = kw
	var max int
	params["List"], max = cache.Ei.PageArticleBE(se, kw, false, false,
		pg, config.Conf.BlogApp.General.PageSize)
	if pg < max {
		vals.Set("page", fmt.Sprint(pg+1))
		params["Next"] = vals.Encode()
	}
	if pg > 1 {
		vals.Set("page", fmt.Sprint(pg-1))
		params["Prev"] = vals.Encode()
	}
	params["PP"] = make(map[int]string, max)
	for i := 0; i < max; i++ {
		vals.Set("page", fmt.Sprint(i+1))
		params["PP"].(map[int]string)[i+1] = vals.Encode()
	}
	params["Cur"] = pg
	renderHTMLAdminLayout(c, "admin-posts", params)
}

// handleAdminSeries 专题列表
func handleAdminSeries(c *gin.Context) {
	params := baseBEParams(c)
	params["Title"] = "专题管理 | " + cache.Ei.Blogger.BTitle
	params["Manage"] = true
	params["Path"] = c.Request.URL.Path
	params["List"] = cache.Ei.Series
	renderHTMLAdminLayout(c, "admin-series", params)
}

// handleAdminSerie 编辑专题
func handleAdminSerie(c *gin.Context) {
	params := baseBEParams(c)

	id, err := strconv.Atoi(c.Query("mid"))
	params["Title"] = "新增专题 | " + cache.Ei.Blogger.BTitle
	if err == nil && id > 0 {
		for _, v := range cache.Ei.Series {
			if v.ID == id {
				params["Title"] = "编辑专题 | " + cache.Ei.Blogger.BTitle
				params["Edit"] = v
				break
			}
		}
	}
	params["Manage"] = true
	params["Path"] = c.Request.URL.Path
	renderHTMLAdminLayout(c, "admin-serie", params)
}

// handleAdminTags 标签列表
func handleAdminTags(c *gin.Context) {
	params := baseBEParams(c)
	params["Title"] = "标签管理 | " + cache.Ei.Blogger.BTitle
	params["Manage"] = true
	params["Path"] = c.Request.URL.Path
	params["List"] = cache.Ei.TagArticles
	renderHTMLAdminLayout(c, "admin-tags", params)
}

// handleDraftDelete 编辑页删除草稿
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

// handleAdminDraft 草稿箱页
func handleAdminDraft(c *gin.Context) {
	params := baseBEParams(c)

	params["Title"] = "草稿箱 | " + cache.Ei.Blogger.BTitle
	params["Manage"] = true
	params["Path"] = c.Request.URL.Path
	var err error
	search := store.SearchArticles{
		Page:   1,
		Limit:  9999,
		Fields: map[string]interface{}{store.SearchArticleDraft: true},
	}
	params["List"], _, err = cache.Ei.LoadArticleList(context.Background(), search)
	if err != nil {
		logrus.Error("handleDraft.LoadDraftArticles: ", err)
		c.Status(http.StatusBadRequest)
	} else {
		c.Status(http.StatusOK)
	}
	renderHTMLAdminLayout(c, "admin-draft", params)
}

// handleAdminTrash 回收箱页
func handleAdminTrash(c *gin.Context) {
	params := baseBEParams(c)
	params["Title"] = "回收箱 | " + cache.Ei.Blogger.BTitle
	params["Manage"] = true
	params["Path"] = c.Request.URL.Path
	var err error
	search := store.SearchArticles{
		Page:   1,
		Limit:  9999,
		Fields: map[string]interface{}{store.SearchArticleTrash: true},
	}
	params["List"], _, err = cache.Ei.LoadArticleList(context.Background(), search)
	if err != nil {
		logrus.Error("handleTrash.LoadArticleList: ", err)
	}
	renderHTMLAdminLayout(c, "admin-trash", params)
}

// handleAdminGeneral 基本设置
func handleAdminGeneral(c *gin.Context) {
	params := baseBEParams(c)
	params["Title"] = "基本设置 | " + cache.Ei.Blogger.BTitle
	params["Setting"] = true
	params["Path"] = c.Request.URL.Path
	renderHTMLAdminLayout(c, "admin-general", params)
}

// handleAdminDiscussion 阅读设置
func handleAdminDiscussion(c *gin.Context) {
	params := baseBEParams(c)
	params["Title"] = "阅读设置 | " + cache.Ei.Blogger.BTitle
	params["Setting"] = true
	params["Path"] = c.Request.URL.Path
	renderHTMLAdminLayout(c, "admin-discussion", params)
}

// renderHTMLAdminLayout 渲染admin页面
func renderHTMLAdminLayout(c *gin.Context, name string, data gin.H) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	// special page
	if name == "login.html" {
		err := htmlTmpl.ExecuteTemplate(c.Writer, name, data)
		if err != nil {
			panic(err)
		}
		return
	}
	buf := bytes.Buffer{}
	err := htmlTmpl.ExecuteTemplate(&buf, name, data)
	if err != nil {
		panic(err)
	}
	data["LayoutContent"] = htemplate.HTML(buf.String())
	err = htmlTmpl.ExecuteTemplate(c.Writer, "adminLayout.html", data)
	if err != nil {
		panic(err)
	}
	if c.Writer.Status() == 0 {
		c.Status(http.StatusOK)
	}
}
