// Package main provides ...
// 这里是前端页面展示相关接口
package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/eiblog/eiblog/setting"
	"github.com/eiblog/utils/logd"
	"github.com/gin-gonic/gin"
)

func Filter() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 过滤黑名单
		if BlackFilter(c) {
			c.Abort()
			return
		}
		// 重定向
		if Redirect(c) {
			c.Abort()
			return
		}
		// 用户cookie，用于统计
		UserCookie(c)
		c.Next()
	}
}

// 黑名单过滤
func BlackFilter(c *gin.Context) bool {
	ip := c.ClientIP()
	if setting.BlackIP[ip] {
		c.String(http.StatusForbidden, "Your IP is blacklisted.")
		return true
	}

	return false
}

// 重定向
func Redirect(c *gin.Context) bool {
	if setting.Conf.Mode.EnableHttps && c.Request.ProtoMajor == 1 {
		var port string
		if strings.Contains(c.Request.Host, ":") {
			port = fmt.Sprintf(":%d", setting.Conf.Mode.HttpsPort)
		}
		c.Redirect(http.StatusMovedPermanently, "https://"+setting.Conf.Mode.Domain+port+c.Request.RequestURI)
		return true
	}

	return false
}

// 用户识别
func UserCookie(c *gin.Context) {
	cookie, err := c.Cookie("u")
	if err != nil || cookie == "" {
		c.SetCookie("u", RandUUIDv4(), 86400*730, "/", "", true, true)
	}
}

// 解析静态文件版本
func StaticVersion(c *gin.Context) (version int) {
	cookie, err := c.Request.Cookie("v")
	if err != nil || cookie.Value != fmt.Sprint(setting.Conf.StaticVersion) {
		return setting.Conf.StaticVersion
	}
	return 0
}

func GetBase() gin.H {
	return gin.H{
		"BlogName": Ei.BlogName,
		"SubTitle": Ei.SubTitle,
		"Twitter":  setting.Conf.Twitter,
		"CopyYear": time.Now().Year(),
		"BTitle":   Ei.BTitle,
		"BeiAn":    Ei.BeiAn,
		"Domain":   setting.Conf.Mode.Domain,
		"Qiniu":    setting.Conf.Qiniu,
		"Disqus":   setting.Conf.Disqus,
	}
}

// not found
func HandleNotFound(c *gin.Context) {
	h := GetBase()
	h["Version"] = StaticVersion(c)
	h["Title"] = "Not Found"
	h["Description"] = "404 Not Found"
	h["Path"] = ""
	c.Status(http.StatusNotFound)
	RenderHTMLFront(c, "notfound", h)
}

// 首页
func HandleHomePage(c *gin.Context) {
	h := GetBase()
	h["Version"] = StaticVersion(c)
	h["Title"] = Ei.BTitle + " | " + Ei.SubTitle
	h["Description"] = "博客首页，" + Ei.SubTitle
	h["Path"] = c.Request.URL.Path
	h["CurrentPage"] = "blog-home"
	pn, err := strconv.Atoi(c.Query("pn"))
	if err != nil || pn < 1 {
		pn = 1
	}
	h["Prev"], h["Next"], h["List"] = PageList(pn, setting.Conf.General.PageNum)
	c.Status(http.StatusOK)
	RenderHTMLFront(c, "home", h)
}

// 专题页
func HandleSeriesPage(c *gin.Context) {
	h := GetBase()
	h["Version"] = StaticVersion(c)
	h["Title"] = "专题 | " + Ei.BTitle
	h["Description"] = "专题列表，" + Ei.SubTitle
	h["Path"] = c.Request.URL.Path
	h["CurrentPage"] = "series"
	h["Article"] = Ei.PageSeries
	c.Status(http.StatusOK)
	RenderHTMLFront(c, "series", h)
}

// 归档页
func HandleArchivesPage(c *gin.Context) {
	h := GetBase()
	h["Version"] = StaticVersion(c)
	h["Title"] = "归档 | " + Ei.BTitle
	h["Description"] = "博客归档，" + Ei.SubTitle
	h["Path"] = c.Request.URL.Path
	h["CurrentPage"] = "archives"
	h["Article"] = Ei.PageArchives
	c.Status(http.StatusOK)
	RenderHTMLFront(c, "archives", h)
}

// 文章
func HandleArticlePage(c *gin.Context) {
	path := c.Param("slug")
	if !strings.HasSuffix(path, ".html") || Ei.MapArticles[path[:len(path)-5]] == nil {
		HandleNotFound(c)
		return
	}
	artc := Ei.MapArticles[path[:len(path)-5]]
	h := GetBase()
	h["Version"] = StaticVersion(c)
	h["Title"] = artc.Title + " | " + Ei.BTitle
	h["Path"] = c.Request.URL.Path
	h["CurrentPage"] = "post-" + artc.Slug
	var name string
	if path == "blogroll.html" {
		name = "blogroll"
		h["Description"] = "友情连接，" + Ei.SubTitle
	} else if path == "about.html" {
		name = "about"
		h["Description"] = "关于作者，" + Ei.SubTitle
	} else {
		h["Description"] = artc.Desc + "，" + Ei.SubTitle
		name = "article"
		h["Copyright"] = Ei.Copyright
		if !artc.UpdateTime.IsZero() {
			h["Days"] = int(time.Now().Sub(artc.UpdateTime).Hours()) / 24
		} else {
			h["Days"] = int(time.Now().Sub(artc.CreateTime).Hours()) / 24
		}
		if artc.SerieID > 0 {
			h["Serie"] = QuerySerie(artc.SerieID)
		}
	}
	h["Article"] = artc
	c.Status(http.StatusOK)
	RenderHTMLFront(c, name, h)
}

// 搜索页
func HandleSearchPage(c *gin.Context) {
	h := GetBase()
	h["Version"] = StaticVersion(c)
	h["Title"] = "站内搜索 | " + Ei.BTitle
	h["Description"] = "站内搜索，" + Ei.SubTitle
	h["Path"] = ""
	h["CurrentPage"] = "search-post"

	q := strings.TrimSpace(c.Query("q"))
	if q != "" {
		start, err := strconv.Atoi(c.Query("start"))
		if start < 1 || err != nil {
			start = 1
		}
		h["Word"] = q

		vals := c.Request.URL.Query()
		result, err := Elasticsearch(q, setting.Conf.General.PageNum, start-1)
		if err != nil {
			logd.Error(err)
		} else {
			result.Took /= 1000
			for i, v := range result.Hits.Hits {
				if artc := Ei.MapArticles[result.Hits.Hits[i].Source.Slug]; len(v.Highlight.Content) == 0 && artc != nil {
					result.Hits.Hits[i].Highlight.Content = []string{artc.Excerpt}
				}
			}
			h["SearchResult"] = result
			if start-setting.Conf.General.PageNum > 0 {
				vals.Set("start", fmt.Sprint(start-setting.Conf.General.PageNum))
				h["Prev"] = vals.Encode()
			}
			if result.Hits.Total >= start+setting.Conf.General.PageNum {
				vals.Set("start", fmt.Sprint(start+setting.Conf.General.PageNum))
				h["Next"] = vals.Encode()
			}
		}
	} else {
		h["HotWords"] = setting.Conf.HotWords
	}
	c.Status(http.StatusOK)
	RenderHTMLFront(c, "search", h)
}

// 评论页
func HandleDisqusFrom(c *gin.Context) {
	params := strings.Split(c.Param("slug"), "|")
	if len(params) != 4 || params[1] == "" {
		c.String(http.StatusOK, "出错啦。。。")
		return
	}
	artc := Ei.MapArticles[params[0]]
	data := gin.H{
		"Title":  "发表评论 | " + Ei.BTitle,
		"ATitle": artc.Title,
		"Thread": params[1],
		"Slug":   artc.Slug,
	}
	err := Tmpl.ExecuteTemplate(c.Writer, "disqus.html", data)
	if err != nil {
		panic(err)
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
}

// feed
func HandleFeed(c *gin.Context) {
	http.ServeFile(c.Writer, c.Request, "static/feed.xml")
}

// opensearch
func HandleOpenSearch(c *gin.Context) {
	http.ServeFile(c.Writer, c.Request, "static/opensearch.xml")
}

// robots
func HandleRobots(c *gin.Context) {
	http.ServeFile(c.Writer, c.Request, "static/robots.txt")
}

// sitemap
func HandleSitemap(c *gin.Context) {
	http.ServeFile(c.Writer, c.Request, "static/sitemap.xml")
}

// cross domain
func HandleCrossDomain(c *gin.Context) {
	http.ServeFile(c.Writer, c.Request, "static/crossdomain.xml")
}

// favicon
func HandleFavicon(c *gin.Context) {
	http.ServeFile(c.Writer, c.Request, "static/favicon.ico")
}

// 服务端推送谷歌统计
func HandleBeacon(c *gin.Context) {
	ua := c.Request.UserAgent()
	// TODO 过滤黑名单
	vals := c.Request.URL.Query()
	vals.Set("v", setting.Conf.Google.V)
	vals.Set("tid", setting.Conf.Google.Tid)
	vals.Set("t", setting.Conf.Google.T)
	cookie, _ := c.Cookie("u")
	vals.Set("cid", cookie)

	vals.Set("dl", c.Request.Referer())
	vals.Set("uip", c.ClientIP())
	go func() {
		req, err := http.NewRequest("POST", setting.Conf.Google.URL, strings.NewReader(vals.Encode()))
		if err != nil {
			logd.Error(err)
			return
		}
		req.Header.Set("User-Agent", ua)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			logd.Error(err)
			return
		}
		defer res.Body.Close()
		data, err := ioutil.ReadAll(res.Body)
		if err != nil {
			logd.Error(err)
			return
		}
		if res.StatusCode/100 != 2 {
			logd.Error(string(data))
		}
	}()
	c.String(http.StatusNoContent, "accepted")
}

// 服务端获取评论详细
type DisqusComments struct {
	ErrNo  int    `json:"errno"`
	ErrMsg string `json:"errmsg"`
	Data   struct {
		Next     string           `json:"next"`
		Total    int              `json:"total"`
		Comments []commentsDetail `json:"comments"`
		Thread   string           `json:"thread"`
	} `json:"data"`
}

type commentsDetail struct {
	Id           string `json:"id"`
	Parent       int    `json:"parent"`
	Name         string `json:"name"`
	Url          string `json:"url"`
	Avatar       string `json:"avatar"`
	CreatedAtStr string `json:"createdAtStr"`
	Message      string `json:"message"`
	IsDeleted    bool   `json:"isDeleted"`
}

func HandleDisqus(c *gin.Context) {
	slug := c.Param("slug")
	cursor := c.Query("cursor")

	dcs := DisqusComments{}
	if artc := Ei.MapArticles[slug]; artc != nil {
		dcs.Data.Thread = artc.Thread
	}
	postsList, err := PostsList(slug, cursor)
	if err != nil {
		logd.Error(err)
		dcs.ErrNo = FAIL
		dcs.ErrMsg = "系统错误"
	} else {
		dcs.ErrNo = postsList.Code
		if postsList.Cursor.HasNext {
			dcs.Data.Next = postsList.Cursor.Next
		}
		dcs.Data.Total = len(postsList.Response)
		dcs.Data.Comments = make([]commentsDetail, len(postsList.Response))
		for i, v := range postsList.Response {
			if dcs.Data.Thread == "" {
				dcs.Data.Thread = v.Thread
			}
			dcs.Data.Comments[i] = commentsDetail{
				Id:           v.Id,
				Name:         v.Author.Name,
				Parent:       v.Parent,
				Url:          v.Author.ProfileUrl,
				Avatar:       v.Author.Avatar.Cache,
				CreatedAtStr: ConvertStr(v.CreatedAt),
				Message:      v.Message,
				IsDeleted:    v.IsDeleted,
			}
		}
	}
	c.JSON(http.StatusOK, dcs)
}

// 发表评论
// [thread:[5279901489] parent:[] identifier:[post-troubleshooting-https]
// next:[] author_name:[你好] author_email:[chenqijing2@163.com] message:[fdsfdsf]]
type DisqusCreate struct {
	ErrNo  int            `json:"errno"`
	ErrMsg string         `json:"errmsg"`
	Data   commentsDetail `json:"data"`
}

func HandleDisqusCreate(c *gin.Context) {
	resp := &DisqusCreate{}
	defer c.JSON(http.StatusOK, resp)

	msg := c.PostForm("message")
	email := c.PostForm("author_email")
	name := c.PostForm("author_name")
	thread := c.PostForm("thread")
	identifier := c.PostForm("identifier")
	if msg == "" || email == "" || name == "" || thread == "" || identifier == "" {
		resp.ErrNo = FAIL
		resp.ErrMsg = "参数错误"
		return
	}
	pc := &PostComment{
		Message:     msg,
		Parent:      c.PostForm("parent"),
		Thread:      thread,
		AuthorEmail: email,
		AuthorName:  name,
		Identifier:  identifier,
		IpAddress:   c.ClientIP(),
	}

	postDetail, err := PostCreate(pc)
	if err != nil {
		logd.Error(err)
		resp.ErrNo = FAIL
		resp.ErrMsg = "系统错误"
		return
	}
	err = PostApprove(postDetail.Response.Id)
	if err != nil {
		logd.Error(err)
		resp.ErrNo = FAIL
		resp.ErrMsg = "系统错误"
		return
	}
	resp.ErrNo = SUCCESS
	resp.Data = commentsDetail{
		Id:           postDetail.Response.Id,
		Name:         name,
		Parent:       postDetail.Response.Parent,
		Url:          postDetail.Response.Author.ProfileUrl,
		Avatar:       postDetail.Response.Author.Avatar.Cache,
		CreatedAtStr: ConvertStr(postDetail.Response.CreatedAt),
		Message:      postDetail.Response.Message,
		IsDeleted:    postDetail.Response.IsDeleted,
	}
}

// 渲染页面
func RenderHTMLFront(c *gin.Context, name string, data gin.H) {
	var buf bytes.Buffer
	err := Tmpl.ExecuteTemplate(&buf, name, data)
	if err != nil {
		panic(err)
	}
	data["LayoutContent"] = template.HTML(buf.String())
	err = Tmpl.ExecuteTemplate(c.Writer, "homeLayout.html", data)
	if err != nil {
		panic(err)
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
}
