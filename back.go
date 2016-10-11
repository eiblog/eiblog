// Package main provides ...
package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/eiblog/eiblog/setting"
	"github.com/eiblog/utils/logd"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2/bson"
)

func isLogin(c *gin.Context) bool {
	session := sessions.Default(c)
	v := session.Get("username")
	if v == nil || v.(string) != Ei.Username {
		return false
	}
	return true
}

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
	c.HTML(http.StatusOK, "login.html", gin.H{
		"BTitle": Ei.BTitle,
	})
}

func HandleLoginPost(c *gin.Context) {
	user := c.PostForm("user")
	pwd := c.PostForm("password")
	// code := c.PostForm("code") // 二次验证
	if user == "" || pwd == "" {
		logd.Info("参数错误", user, pwd)
		c.Redirect(http.StatusFound, "/admin/login")
		return
	}
	if Ei.Username != user || !VerifyPasswd(Ei.Password, user, pwd) {
		logd.Info("账号或密码错误", user, pwd)
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
	return gin.H{"Author": Ei.Username}
}

// 个人配置
func HandleProfile(c *gin.Context) {
	h := GetBack()
	h["Console"] = true
	h["Path"] = c.Request.URL.Path
	h["Title"] = "个人配置 | " + Ei.BTitle
	h["Account"] = Ei
	h["Profile"] = true
	c.HTML(http.StatusOK, "backLayout.html", h)
}

// 写文章==>Write
func HandlePost(c *gin.Context) {
	h := GetBack()
	id, err := strconv.Atoi(c.Query("cid"))
	if artc := QueryArticle(int32(id)); err == nil && id > 0 && artc != nil {
		h["Title"] = "编辑文章 | " + Ei.BTitle
		h["Edit"] = artc
	} else {
		h["Title"] = "撰写文章 | " + Ei.BTitle
	}
	h["Post"] = true
	h["Path"] = c.Request.URL.Path
	h["Title"] = "撰写文章 | " + Ei.BTitle
	h["Domain"] = setting.Conf.Mode.Domain
	h["Series"] = Ei.Series
	c.HTML(http.StatusOK, "backLayout.html", h)
}

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
	c.JSON(http.StatusOK, nil)
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
	h["Posts"] = true
	h["Series"] = Ei.Series
	h["Serie"] = se
	h["KW"] = kw
	var max int
	max, h["List"] = PageListBack(se, kw, false, false, pg, setting.Conf.PageSize)
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
	c.HTML(http.StatusOK, "backLayout.html", h)
}

// 专题列表
func HandleSeries(c *gin.Context) {
	h := GetBack()
	h["Manage"] = true
	h["Path"] = c.Request.URL.Path
	h["Title"] = "专题管理 | " + Ei.BTitle
	h["Series"] = true
	h["List"] = Ei.Series
	c.HTML(http.StatusOK, "backLayout.html", h)
}

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
	h["Serie"] = true
	c.HTML(http.StatusOK, "backLayout.html", h)
}

// 标签列表
func HandleTags(c *gin.Context) {
	h := GetBack()
	h["Manage"] = true
	h["Path"] = c.Request.URL.Path
	h["Title"] = "标签管理 | " + Ei.BTitle
	h["Tags"] = true
	h["List"] = Ei.Tags
	c.HTML(http.StatusOK, "backLayout.html", h)
}

// 草稿箱
func HandleDraft(c *gin.Context) {
	h := GetBack()
	h["Manage"] = true
	h["Path"] = c.Request.URL.Path
	h["Title"] = "草稿箱 | " + Ei.BTitle
	h["Draft"] = true
	var err error
	h["List"], err = LoadDraft()
	if err != nil {
		logd.Error(err)
		c.HTML(http.StatusBadRequest, "backLayout.html", h)
		return
	}
	c.HTML(http.StatusOK, "backLayout.html", h)
}

// 回收箱
func HandleTrash(c *gin.Context) {
	h := GetBack()
	h["Manage"] = true
	h["Path"] = c.Request.URL.Path
	h["Title"] = "回收箱 | " + Ei.BTitle
	h["Trash"] = true
	var err error
	h["List"], err = LoadTrash()
	if err != nil {
		logd.Error(err)
		c.HTML(http.StatusBadRequest, "backLayout.html", h)
		return
	}
	c.HTML(http.StatusOK, "backLayout.html", h)
}

// 基本设置==>Setting
func HandleGeneral(c *gin.Context) {
	h := GetBack()
	h["Setting"] = true
	h["Path"] = c.Request.URL.Path
	h["Title"] = "基本设置 | " + Ei.BTitle
	h["General"] = true
	c.HTML(http.StatusOK, "backLayout.html", h)
}

// 阅读设置
func HandleDiscussion(c *gin.Context) {
	h := GetBack()
	h["Setting"] = true
	h["Path"] = c.Request.URL.Path
	h["Title"] = "阅读设置 | " + Ei.BTitle
	h["Discussion"] = true
	c.HTML(http.StatusOK, "backLayout.html", h)
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
