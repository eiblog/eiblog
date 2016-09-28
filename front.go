// Package main provides ...
// 这里是前端页面展示相关接口
package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/EiBlog/eiblog/setting"
	"github.com/EiBlog/utils/logd"
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

func HandleNotFound(c *gin.Context) {
	h := GetBase()
	h["Version"] = StaticVersion(c)
	h["Title"] = "Not Found"
	h["NotFoundPage"] = true
	h["Path"] = ""
	c.HTML(http.StatusNotFound, "homeLayout.html", h)
}

func GetBase() gin.H {
	return gin.H{
		"Favicon":  setting.Conf.Favicon,
		"BlogName": Ei.BlogName,
		"SubTitle": Ei.SubTitle,
		"Twitter":  setting.Conf.Twitter,
		"RSS":      setting.Conf.RSS,
		"Search":   setting.Conf.Search,
		"CopyYear": time.Now().Year(),
		"BTitle":   Ei.BTitle,
		"BeiAn":    Ei.BeiAn,
		"Domain":   runmode.Domain,
		"Static":   setting.Conf.Static,
	}
}

func HandleHomePage(c *gin.Context) {
	h := GetBase()
	h["Version"] = StaticVersion(c)
	h["Title"] = Ei.BTitle
	h["Path"] = c.Request.URL.Path
	h["HomePage"] = true
	h["CurrentPage"] = "blog-home"
	pn, err := strconv.Atoi(c.Query("pn"))
	if err != nil || pn < 1 {
		pn = 1
	}
	h["Prev"], h["Next"], h["List"] = PageList(pn, setting.Conf.PageNum)
	c.HTML(http.StatusOK, "homeLayout.html", h)
}

func HandleSeriesPage(c *gin.Context) {
	h := GetBase()
	h["Version"] = StaticVersion(c)
	h["Title"] = "专题 | " + Ei.BTitle
	h["Path"] = c.Request.URL.Path
	h["SeriesPage"] = true
	h["CurrentPage"] = "series"
	h["Article"] = Ei.PageSeries
	c.HTML(http.StatusOK, "homeLayout.html", h)
}

func HandleArchivesPage(c *gin.Context) {
	h := GetBase()
	h["Version"] = StaticVersion(c)
	h["Title"] = "归档 | " + Ei.BTitle
	h["Path"] = c.Request.URL.Path
	h["ArchivesPage"] = true
	h["CurrentPage"] = "archives"
	h["Article"] = Ei.PageArchives
	c.HTML(http.StatusOK, "homeLayout.html", h)
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
	if path == "blogroll.html" {
		h["BlogrollPage"] = true
	} else if path == "about.html" {
		h["AboutPage"] = true
	} else {
		h["ArticlePage"] = true
		h["Copyright"] = Ei.Copyright
		if !artc.UpdateTime.IsZero() {
			h["Days"] = int(time.Now().Sub(artc.UpdateTime).Hours()) / 24
		}
		if artc.SerieID > 0 {
			h["Serie"] = QuerySerie(artc.SerieID)
		}
	}
	h["Article"] = artc
	h["EnableHttps"] = runmode.EnableHttps
	c.HTML(http.StatusOK, "homeLayout.html", h)
}

type temp struct {
	Title      string
	Slug       string
	URL        string
	Img        string
	CreateTime time.Time
}

func HandleSearchPage(c *gin.Context) {
	h := GetBase()
	h["Version"] = StaticVersion(c)
	h["Title"] = "站内搜索 | " + Ei.BTitle
	h["Path"] = ""
	h["SearchPage"] = true
	h["CurrentPage"] = "search-post"

	q := c.Query("q")
	start, err := strconv.Atoi(c.Query("start"))
	if start < 1 || err != nil {
		start = 1
	}
	if q != "" {
		h["Word"] = q
		// TODO search
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
	c.HTML(http.StatusOK, "homeLayout.html", h)
}

func HandleFeed(c *gin.Context) {
	http.ServeFile(c.Writer, c.Request, "conf/feed.xml")
}

func HandleOpenSearch(c *gin.Context) {
	c.Header("Content-Type", "application/xml; charset=utf-8")
	c.Writer.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	c.HTML(http.StatusOK, "opensearch.xml", gin.H{
		"BTitle":   Ei.BTitle,
		"SubTitle": Ei.SubTitle,
	})
}

func HandleRobots(c *gin.Context) {
	http.ServeFile(c.Writer, c.Request, "conf/robots.txt")
}

func HandleSitemap(c *gin.Context) {
	http.ServeFile(c.Writer, c.Request, "conf/sitemap.xml")
}

func HandleComments(c *gin.Context) {
	// TODO comments
	var ss = map[string]interface{}{
		"errno":  0,
		"errmsg": "",
		"data": map[string]interface{}{
			"next":       "",
			"commentNum": 2,
			"comments": []map[string]interface{}{
				map[string]interface{}{
					"id":           "2361914870",
					"name":         "Rekey Luo",
					"url":          "https://disqus.com/by/rekeyluo/",
					"avatar":       "//a.disquscdn.com/uploads/users/15860/7550/avatar92.jpg?1438917750",
					"createdAt":    "2015-11-16T05:00:02",
					"createdAtStr": "9 months ago",
					"message":      "你最近对 http2 ssl 相关关注好多啊。",
					"children": []map[string]interface{}{
						map[string]interface{}{
							"id":           "2361915528",
							"name":         "Jerry Qu",
							"url":          "https://disqus.com/by/JerryQu/",
							"avatar":       "//a.disquscdn.com/uploads/users/1668/8837/avatar92.jpg?1472281172",
							"createdAt":    "2015-11-16T05:01:05",
							"createdAtStr": "9 months ago",
							"message":      "嗯，最近对 web 性能优化这一块研究得比较多。",
						},
					},
				},
			},
		},
	}
	c.JSON(http.StatusOK, ss)
}
