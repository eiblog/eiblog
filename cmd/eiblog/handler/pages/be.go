package pages

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"image/png"
	"net/http"
	"strconv"

	"github.com/eiblog/eiblog/cmd/eiblog/config"
	"github.com/eiblog/eiblog/cmd/eiblog/handler/internal"
	"github.com/eiblog/eiblog/cmd/eiblog/handler/internal/store"
	"github.com/eiblog/eiblog/pkg/middleware"
	"github.com/pquerna/otp/totp"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// baseBEParams 基础参数
func baseBEParams(_ *gin.Context) gin.H {
	return gin.H{
		"Author": internal.Ei.Account.Username,
		"Qiniu":  config.Conf.Qiniu,
	}
}

// handleLoginPage 登录页面
func handleLoginPage(c *gin.Context) {
	logout := c.Query("logout")
	if logout == "true" {
		middleware.SetLogout(c)
	} else if middleware.IsLogined(c) {
		c.Redirect(http.StatusFound, "/admin/profile")
		return
	}
	params := gin.H{
		"BTitle": internal.Ei.Blogger.BTitle,
		"TwoFactor": config.Conf.General.TwoFactor &&
			internal.Ei.Account.TwoFactorSecret != "",
	}
	renderHTMLAdminLayout(c, "login.html", params)
}

// handleAdminProfile 个人配置
func handleAdminProfile(c *gin.Context) {
	params := baseBEParams(c)
	params["Title"] = "个人配置 | " + internal.Ei.Blogger.BTitle
	params["Path"] = c.Request.URL.Path
	params["Console"] = true
	params["Ei"] = internal.Ei
	if c.Query("unbind") == "true" {
		internal.Ei.Account.TwoFactorSecret = ""
		_ = internal.Store.UpdateAccount(context.Background(), internal.Ei.Account.Username,
			map[string]interface{}{
				"two_factor_secret": "",
			})
	}
	if config.Conf.General.TwoFactor &&
		internal.Ei.Account.TwoFactorSecret == "" {
		key, err := totp.Generate(totp.GenerateOpts{
			Issuer:      config.Conf.Host,
			AccountName: internal.Ei.Account.Username,
		})
		internal.TwoFactorSecret = key.Secret()
		if err == nil {
			var buf bytes.Buffer
			img, err := key.Image(200, 200)
			if err != nil {
				panic(err)
			}
			png.Encode(&buf, img)
			b64 := base64.StdEncoding.EncodeToString(buf.Bytes())
			params["TwoFactorSecret"] = "data:image/png;base64," + b64
		}
	}
	renderHTMLAdminLayout(c, "admin-profile", params)
}

// T tag struct
type T struct {
	ID   string `json:"id"`
	Tags string `json:"tags"`
}

// handleAdminPost 写文章页
func handleAdminPost(c *gin.Context) {
	params := baseBEParams(c)
	id, err := strconv.Atoi(c.Query("cid"))
	if err == nil && id > 0 {
		article, _ := internal.Store.LoadArticle(context.Background(), id)
		if article != nil {
			params["Title"] = "编辑文章 | " + internal.Ei.Blogger.BTitle
			params["Edit"] = article
		}
	}
	if params["Title"] == nil {
		params["Title"] = "撰写文章 | " + internal.Ei.Blogger.BTitle
	}
	params["Path"] = c.Request.URL.Path
	params["Domain"] = config.Conf.Host
	params["Series"] = internal.Ei.Series
	var tags []T
	for tag := range internal.Ei.TagArticles {
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
	params["Title"] = "文章管理 | " + internal.Ei.Blogger.BTitle
	params["Manage"] = true
	params["Path"] = c.Request.URL.Path
	params["Series"] = internal.Ei.Series
	params["Serie"] = se
	params["KW"] = kw
	var max int
	params["List"], max = internal.Ei.PageArticleBE(se, kw, false, false,
		pg, config.Conf.General.PageSize)
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
	params["Title"] = "专题管理 | " + internal.Ei.Blogger.BTitle
	params["Manage"] = true
	params["Path"] = c.Request.URL.Path
	params["List"] = internal.Ei.Series
	renderHTMLAdminLayout(c, "admin-series", params)
}

// handleAdminSerie 编辑专题
func handleAdminSerie(c *gin.Context) {
	params := baseBEParams(c)

	id, err := strconv.Atoi(c.Query("mid"))
	params["Title"] = "新增专题 | " + internal.Ei.Blogger.BTitle
	if err == nil && id > 0 {
		for _, v := range internal.Ei.Series {
			if v.ID == id {
				params["Title"] = "编辑专题 | " + internal.Ei.Blogger.BTitle
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
	params["Title"] = "标签管理 | " + internal.Ei.Blogger.BTitle
	params["Manage"] = true
	params["Path"] = c.Request.URL.Path
	params["List"] = internal.Ei.TagArticles
	renderHTMLAdminLayout(c, "admin-tags", params)
}

// handleDraftDelete 编辑页删除草稿
func handleDraftDelete(c *gin.Context) {
	id, err := strconv.Atoi(c.Query("cid"))
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	err = internal.Store.RemoveArticle(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "删除错误"})
		return
	}
	c.Redirect(http.StatusFound, "/admin/write-post")
}

// handleAdminDraft 草稿箱页
func handleAdminDraft(c *gin.Context) {
	params := baseBEParams(c)

	params["Title"] = "草稿箱 | " + internal.Ei.Blogger.BTitle
	params["Manage"] = true
	params["Path"] = c.Request.URL.Path
	var err error
	search := store.SearchArticles{
		Page:   1,
		Limit:  9999,
		Fields: map[string]interface{}{store.SearchArticleDraft: true},
	}
	params["List"], _, err = internal.Store.LoadArticleList(context.Background(), search)
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
	params["Title"] = "回收箱 | " + internal.Ei.Blogger.BTitle
	params["Manage"] = true
	params["Path"] = c.Request.URL.Path
	var err error
	search := store.SearchArticles{
		Page:   1,
		Limit:  9999,
		Fields: map[string]interface{}{store.SearchArticleTrash: true},
	}
	params["List"], _, err = internal.Store.LoadArticleList(context.Background(), search)
	if err != nil {
		logrus.Error("handleTrash.LoadArticleList: ", err)
	}
	renderHTMLAdminLayout(c, "admin-trash", params)
}

// handleAdminGeneral 基本设置
func handleAdminGeneral(c *gin.Context) {
	params := baseBEParams(c)
	params["Title"] = "基本设置 | " + internal.Ei.Blogger.BTitle
	params["Setting"] = true
	params["Path"] = c.Request.URL.Path
	renderHTMLAdminLayout(c, "admin-general", params)
}

// handleAdminDiscussion 阅读设置
func handleAdminDiscussion(c *gin.Context) {
	params := baseBEParams(c)
	params["Title"] = "阅读设置 | " + internal.Ei.Blogger.BTitle
	params["Setting"] = true
	params["Path"] = c.Request.URL.Path
	renderHTMLAdminLayout(c, "admin-discussion", params)
}

// renderHTMLAdminLayout 渲染admin页面
func renderHTMLAdminLayout(c *gin.Context, name string, data gin.H) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	// special page
	if name == "login.html" {
		err := internal.HTMLTemplate.ExecuteTemplate(c.Writer, name, data)
		if err != nil {
			panic(err)
		}
		return
	}
	buf := bytes.Buffer{}
	err := internal.HTMLTemplate.ExecuteTemplate(&buf, name, data)
	if err != nil {
		panic(err)
	}
	data["LayoutContent"] = template.HTML(buf.String())
	err = internal.HTMLTemplate.ExecuteTemplate(c.Writer, "adminLayout.html", data)
	if err != nil {
		panic(err)
	}
	if c.Writer.Status() == 0 {
		c.Status(http.StatusOK)
	}
}
