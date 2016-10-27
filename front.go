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

	q := strings.TrimSpace(c.Query("q"))
	if q != "" {
		start, err := strconv.Atoi(c.Query("start"))
		if start < 1 || err != nil {
			start = 1
		}
		h["Word"] = q
		var result *ESSearchResult
		vals := c.Request.URL.Query()
		result = Elasticsearch(q, setting.Conf.PageNum, start-1)
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
	} else {
		h["HotWords"] = setting.Conf.HotWords
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
type DisqusComments struct {
	ErrNo  int    `json:"errno"`
	ErrMsg string `json:"errmsg"`
	Data   struct {
		Next     string           `json:"next"`
		Total    int              `json:"total,omitempty"`
		Comments []commentsDetail `json:"comments"`
	} `json:"data"`
}

type commentsDetail struct {
	Id           string `json:"id"`
	Parent       int    `json:"parent"`
	Name         string `json:"name"`
	Url          string `json:"url"`
	Avatar       string `json:"avatar"`
	CreatedAt    string `json:"createdAt"`
	CreatedAtStr string `json:"createdAtStr"`
	Message      string `json:"message"`
}

func HandleDisqus(c *gin.Context) {
	slug := c.Param("slug")
	cursor := c.Query("cursor")
	dcs := DisqusComments{}
	postsList := PostsList(slug, cursor)
	if postsList != nil {
		dcs.ErrNo = postsList.Code
		if postsList.Cursor.HasNext {
			dcs.Data.Next = postsList.Cursor.Next
		}
		if cursor == "" {
			dcs.Data.Total = Ei.MapArticles[slug].Count
		}
		dcs.Data.Comments = make([]commentsDetail, len(postsList.Response))
		for i, v := range postsList.Response {
			dcs.Data.Comments[i] = commentsDetail{
				Id:           v.Id,
				Name:         v.Author.Name,
				Parent:       v.Parent,
				Url:          v.Author.ProfileUrl,
				Avatar:       v.Author.Avatar.Cache,
				CreatedAt:    v.CreatedAt,
				CreatedAtStr: ConvertStr(v.CreatedAt),
				Message:      v.Message,
			}
		}
	} else {
		dcs.ErrNo = FAIL
		dcs.ErrMsg = "系统错误"
	}
	c.JSON(http.StatusOK, dcs)
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
