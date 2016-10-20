// Package main provides ...
// 这里是前端页面展示相关接口
package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
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
		BlackFilter(c)
		// 用户cookie，用于统计
		UserCookie(c)
		c.Next()
	}
}

// 用户识别
func UserCookie(c *gin.Context) {
	cookie, err := c.Request.Cookie("u")
	if err != nil || cookie.Value == "" {
		// TODO cookie操作
		b := []byte(c.ClientIP() + time.Now().String())
		c.SetCookie("u", fmt.Sprintf("%x", SHA1(b)), 86400*999, "/", "", true, true)
	}
}

// 黑名单过滤
func BlackFilter(c *gin.Context) {
	ip := c.ClientIP()
	if setting.BlackIP[ip] {
		c.Abort()
		c.String(http.StatusForbidden, "Your IP is blacklisted.")
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
		"Favicon":  setting.Conf.Favicon,
		"BlogName": Ei.BlogName,
		"SubTitle": Ei.SubTitle,
		"Twitter":  setting.Conf.Twitter,
		"CopyYear": time.Now().Year(),
		"BTitle":   Ei.BTitle,
		"BeiAn":    Ei.BeiAn,
		"Domain":   setting.Conf.Mode.Domain,
		"Static":   setting.Conf.Static,
	}
}

func HandleNotFound(c *gin.Context) {
	h := GetBase()
	h["Version"] = StaticVersion(c)
	h["Title"] = "Not Found"
	h["Path"] = ""
	c.Status(http.StatusNotFound)
	RenderHTMLFront(c, "notfound", h)
}

func HandleHomePage(c *gin.Context) {
	h := GetBase()
	h["Version"] = StaticVersion(c)
	h["Title"] = Ei.BTitle + " | " + Ei.SubTitle
	h["Path"] = c.Request.URL.Path
	h["CurrentPage"] = "blog-home"
	pn, err := strconv.Atoi(c.Query("pn"))
	if err != nil || pn < 1 {
		pn = 1
	}
	h["Prev"], h["Next"], h["List"] = PageList(pn, setting.Conf.PageNum)
	c.Status(http.StatusOK)
	RenderHTMLFront(c, "home", h)
}

func HandleSeriesPage(c *gin.Context) {
	h := GetBase()
	h["Version"] = StaticVersion(c)
	h["Title"] = "专题 | " + Ei.BTitle
	h["Path"] = c.Request.URL.Path
	h["CurrentPage"] = "series"
	h["Article"] = Ei.PageSeries
	c.Status(http.StatusOK)
	RenderHTMLFront(c, "series", h)
}

func HandleArchivesPage(c *gin.Context) {
	h := GetBase()
	h["Version"] = StaticVersion(c)
	h["Title"] = "归档 | " + Ei.BTitle
	h["Path"] = c.Request.URL.Path
	h["CurrentPage"] = "archives"
	h["Article"] = Ei.PageArchives
	c.Status(http.StatusOK)
	RenderHTMLFront(c, "archives", h)
}

func HandleArticlePage(c *gin.Context) {
	path := c.Param("slug")
	artc := Ei.MapArticles[path[0:strings.Index(path, ".")]]
	if artc == nil {
		HandleNotFound(c)
		return
	}
	h := GetBase()
	h["Version"] = StaticVersion(c)
	h["Title"] = artc.Title + " | " + Ei.BTitle
	h["Path"] = c.Request.URL.Path
	h["CurrentPage"] = "post-" + artc.Slug
	var name string
	if path == "blogroll.html" {
		name = "blogroll"
	} else if path == "about.html" {
		name = "about"
	} else {
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

func HandleSearchPage(c *gin.Context) {
	h := GetBase()
	h["Version"] = StaticVersion(c)
	h["Title"] = "站内搜索 | " + Ei.BTitle
	h["Path"] = ""
	h["CurrentPage"] = "search-post"

	q := c.Query("q")
	start, err := strconv.Atoi(c.Query("start"))
	if start < 1 || err != nil {
		start = 1
	}
	if q != "" {
		h["Word"] = q
		h["HotWords"] = []string{"docker"}
		var result *ESSearchResult
		vals := c.Request.URL.Query()
		reg := regexp.MustCompile(`^[a-z]+:\w+$`)
		logd.Debug(reg.MatchString(q))
		if reg.MatchString(q) {
			result = ElasticsearchSimple(q, setting.Conf.PageNum, start-1)
		} else {
			result = Elasticsearch(q, setting.Conf.PageNum, start-1)
		}
		if result != nil {
			result.Took /= 1000
			for i, v := range result.Hits.Hits {
				if artc := Ei.MapArticles[result.Hits.Hits[i].Source.Slug]; len(v.Highlight.Content) == 0 && artc != nil {
					result.Hits.Hits[i].Highlight.Content = []string{artc.Excerpt}
				}
			}
			h["SearchResult"] = result
			if start-setting.Conf.PageNum > 0 {
				vals.Set("start", fmt.Sprint(start-setting.Conf.PageNum))
				h["Prev"] = vals.Encode()
			}
			if result.Hits.Total >= start+setting.Conf.PageNum {
				vals.Set("start", fmt.Sprint(start+setting.Conf.PageNum))
				h["Next"] = vals.Encode()
			}
		}
	}
	c.Status(http.StatusOK)
	RenderHTMLFront(c, "search", h)
}

func HandleFeed(c *gin.Context) {
	http.ServeFile(c.Writer, c.Request, "static/feed.xml")
}

func HandleOpenSearch(c *gin.Context) {
	http.ServeFile(c.Writer, c.Request, "static/opensearch.xml")
}

func HandleRobots(c *gin.Context) {
	http.ServeFile(c.Writer, c.Request, "static/robots.txt")
}

func HandleSitemap(c *gin.Context) {
	http.ServeFile(c.Writer, c.Request, "static/sitemap.xml")
}

// 服务端推送谷歌统计
func HandleBeacon(c *gin.Context) {
	ua := c.Request.UserAgent()
	// TODO 过滤黑名单
	go func() {
		req, err := http.NewRequest("POST", "https://www.google-analytics.com/collect", strings.NewReader(c.Request.URL.RawQuery))
		if err != nil {
			logd.Error(err)
			return
		}
		req.Header.Set("User-Agent", ua)
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
	c.String(http.StatusAccepted, "accepted")
}

// 服务端获取评论详细
func HandleDisqus(c *gin.Context) {
	slug := c.Query("slug")
	logd.Debug(slug)
	// TODO comments
	var ss = map[string]interface{}{
		"errno":  0,
		"errmsg": "",
		"data": map[string]interface{}{
			"next":  "",
			"total": 3,
			"comments": []map[string]interface{}{
				map[string]interface{}{
					"id":           "2361914870",
					"name":         "Rekey Luo",
					"parent":       0,
					"url":          "https://disqus.com/by/rekeyluo/",
					"avatar":       "//a.disquscdn.com/uploads/users/15860/7550/avatar92.jpg?1438917750",
					"createdAt":    "2015-11-16T05:00:02",
					"createdAtStr": "9 months ago",
					"message":      "你最近对 http2 ssl 相关关注好多啊。",
				},
				map[string]interface{}{
					"id":           "2361915528",
					"name":         "Jerry Qu",
					"parent":       0,
					"url":          "https://disqus.com/by/JerryQu/",
					"avatar":       "//a.disquscdn.com/uploads/users/1668/8837/avatar92.jpg?1472281172",
					"createdAt":    "2015-11-16T05:01:05",
					"createdAtStr": "9 months ago",
					"message":      "嗯，最近对 web 性能优化这一块研究得比较多。",
				},
			},
		},
	}
	c.JSON(http.StatusOK, ss)
}

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
