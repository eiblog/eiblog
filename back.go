// Package main provides ...
package main

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/eiblog/eiblog/setting"
	"github.com/eiblog/utils/logd"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
)

// 是否登录
func isLogin(c *gin.Context) bool {
	session := sessions.Default(c)
	v := session.Get("username")
	if v == nil || v.(string) != Ei.Username {
		return false
	}
	return true
}

// 登陆过滤
func AuthFilter() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isLogin(c) {
			c.Abort()
			c.Redirect(http.StatusFound, "/admin/login")
			return
		}
		c.Next()
	}
}

// 登录界面
func HandleLogin(c *gin.Context) {
	logout := c.Query("logout")
	if logout == "true" {
		session := sessions.Default(c)
		session.Delete("username")
		session.Save()
	} else if isLogin(c) {
		c.Redirect(http.StatusFound, "/admin/profile")
		return
	}
	c.Status(http.StatusOK)
	RenderHTMLBack(c, "login.html", gin.H{"BTitle": Ei.BTitle})
}

// 登陆接口
func HandleLoginPost(c *gin.Context) {
	user := c.PostForm("user")
	pwd := c.PostForm("password")
	// code := c.PostForm("code") // 二次验证
	if user == "" || pwd == "" {
		logd.Print("参数错误", user, pwd)
		c.Redirect(http.StatusFound, "/admin/login")
		return
	}
	if Ei.Username != user || !VerifyPasswd(Ei.Password, user, pwd) {
		logd.Printf("账号或密码错误 %s, %s\n", user, pwd)
		c.Redirect(http.StatusFound, "/admin/login")
		return
	}
	session := sessions.Default(c)
	session.Set("username", user)
	session.Save()
	Ei.LoginIP = c.ClientIP()
	Ei.LoginTime = time.Now()
	UpdateAccountField(bson.M{"$set": bson.M{"loginip": Ei.LoginIP, "logintime": Ei.LoginTime}})
	c.Redirect(http.StatusFound, "/admin/profile")
}

func GetBack() gin.H {
	return gin.H{"Author": Ei.Username, "Qiniu": setting.Conf.Qiniu}
}

// 个人配置
func HandleProfile(c *gin.Context) {
	h := GetBack()
	h["Console"] = true
	h["Path"] = c.Request.URL.Path
	h["Title"] = "个人配置 | " + Ei.BTitle
	h["Account"] = Ei
	c.Status(http.StatusOK)
	RenderHTMLBack(c, "admin-profile", h)
}

// 写文章==>Write
type T struct {
	ID   string `json:"id"`
	Tags string `json:"tags"`
}

func HandlePost(c *gin.Context) {
	h := GetBack()
	id, err := strconv.Atoi(c.Query("cid"))
	if err == nil && id > 0 {
		artc := QueryArticle(int32(id))
		if artc != nil {
			h["Title"] = "编辑文章 | " + Ei.BTitle
			h["Edit"] = artc
		}
	}
	if h["Title"] == nil {
		h["Title"] = "撰写文章 | " + Ei.BTitle
	}
	h["Path"] = c.Request.URL.Path
	h["Domain"] = setting.Conf.Mode.Domain
	h["Series"] = Ei.Series
	var tags []T
	for tag, _ := range Ei.Tags {
		tags = append(tags, T{tag, tag})
	}
	h["Tags"] = tags
	c.Status(http.StatusOK)
	RenderHTMLBack(c, "admin-post", h)
}

// 删除草稿
func HandleDraftDelete(c *gin.Context) {
	id, err := strconv.Atoi(c.Query("cid"))
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	if err = RemoveArticle(int32(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "删除错误"})
		return
	}
	c.Redirect(http.StatusFound, "/admin/write-post")
}

// 文章管理==>Manage
func HandlePosts(c *gin.Context) {
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
	h := GetBack()
	h["Manage"] = true
	h["Path"] = c.Request.URL.Path
	h["Title"] = "文章管理 | " + Ei.BTitle
	h["Series"] = Ei.Series
	h["Serie"] = se
	h["KW"] = kw
	var max int
	max, h["List"] = PageListBack(se, kw, false, false, pg, setting.Conf.General.PageSize)
	if pg < max {
		vals.Set("page", fmt.Sprint(pg+1))
		h["Next"] = vals.Encode()
	}
	if pg > 1 {
		vals.Set("page", fmt.Sprint(pg-1))
		h["Prev"] = vals.Encode()
	}
	h["PP"] = make(map[int]string, max)
	for i := 0; i < max; i++ {
		vals.Set("page", fmt.Sprint(i+1))
		h["PP"].(map[int]string)[i+1] = vals.Encode()
	}
	h["Cur"] = pg
	c.Status(http.StatusOK)
	RenderHTMLBack(c, "admin-posts", h)
}

// 专题列表
func HandleSeries(c *gin.Context) {
	h := GetBack()
	h["Manage"] = true
	h["Path"] = c.Request.URL.Path
	h["Title"] = "专题管理 | " + Ei.BTitle
	h["List"] = Ei.Series
	c.Status(http.StatusOK)
	RenderHTMLBack(c, "admin-series", h)
}

// 编辑专题
func HandleSerie(c *gin.Context) {
	h := GetBack()
	id, err := strconv.Atoi(c.Query("mid"))
	if serie := QuerySerie(int32(id)); err == nil && id > 0 && serie != nil {
		h["Title"] = "编辑专题 | " + Ei.BTitle
		h["Edit"] = serie
	} else {
		h["Title"] = "新增专题 | " + Ei.BTitle
	}
	h["Manage"] = true
	h["Path"] = c.Request.URL.Path
	c.Status(http.StatusOK)
	RenderHTMLBack(c, "admin-serie", h)
}

// 标签列表
func HandleTags(c *gin.Context) {
	h := GetBack()
	h["Manage"] = true
	h["Path"] = c.Request.URL.Path
	h["Title"] = "标签管理 | " + Ei.BTitle
	h["List"] = Ei.Tags
	c.Status(http.StatusOK)
	RenderHTMLBack(c, "admin-tags", h)
}

// 草稿箱
func HandleDraft(c *gin.Context) {
	h := GetBack()
	h["Manage"] = true
	h["Path"] = c.Request.URL.Path
	h["Title"] = "草稿箱 | " + Ei.BTitle
	var err error
	h["List"], err = LoadDraft()
	if err != nil {
		logd.Error(err)
		c.Status(http.StatusBadRequest)
	} else {
		c.Status(http.StatusOK)
	}
	RenderHTMLBack(c, "admin-draft", h)
}

// 回收箱
func HandleTrash(c *gin.Context) {
	h := GetBack()
	h["Manage"] = true
	h["Path"] = c.Request.URL.Path
	h["Title"] = "回收箱 | " + Ei.BTitle
	var err error
	h["List"], err = LoadTrash()
	if err != nil {
		logd.Error(err)
		c.HTML(http.StatusBadRequest, "backLayout.html", h)
		return
	}
	c.Status(http.StatusOK)
	RenderHTMLBack(c, "admin-trash", h)
}

// 基本设置==>Setting
func HandleGeneral(c *gin.Context) {
	h := GetBack()
	h["Setting"] = true
	h["Path"] = c.Request.URL.Path
	h["Title"] = "基本设置 | " + Ei.BTitle
	c.Status(http.StatusOK)
	RenderHTMLBack(c, "admin-general", h)
}

// 阅读设置
func HandleDiscussion(c *gin.Context) {
	h := GetBack()
	h["Setting"] = true
	h["Path"] = c.Request.URL.Path
	h["Title"] = "阅读设置 | " + Ei.BTitle
	c.Status(http.StatusOK)
	RenderHTMLBack(c, "admin-discussion", h)
}

// api
func HandleAPI(c *gin.Context) {
	action := c.Param("action")
	logd.Debug("action=======>", action)
	api := APIs[action]
	if api == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid API Request"})
		return
	}
	api(c)
}

// 渲染 html
func RenderHTMLBack(c *gin.Context, name string, data gin.H) {
	if name == "login.html" {
		err := Tmpl.ExecuteTemplate(c.Writer, name, data)
		if err != nil {
			panic(err)
		}
		c.Header("Content-Type", "text/html; charset=utf-8")
		return
	}
	var buf bytes.Buffer
	err := Tmpl.ExecuteTemplate(&buf, name, data)
	if err != nil {
		panic(err)
	}
	data["LayoutContent"] = template.HTML(buf.String())
	err = Tmpl.ExecuteTemplate(c.Writer, "backLayout.html", data)
	if err != nil {
		panic(err)
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
}
