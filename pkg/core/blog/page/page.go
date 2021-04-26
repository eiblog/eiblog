// Package page provides ...
package page

import (
	"bytes"
	"fmt"
	htemplate "html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/eiblog/eiblog/pkg/cache"
	"github.com/eiblog/eiblog/pkg/config"
	"github.com/eiblog/eiblog/pkg/internal"
	"github.com/eiblog/eiblog/tools"

	"github.com/eiblog/utils/tmpl"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// htmlTmpl html template cache
var htmlTmpl *template.Template

func init() {
	htmlTmpl = template.New("eiblog").Funcs(tmpl.TplFuncMap)
	root := filepath.Join(config.WorkDir, "website")
	files := tools.ReadDirFiles(root, func(name string) bool {
		if name == ".DS_Store" {
			return true
		}
		return false
	})
	_, err := htmlTmpl.ParseFiles(files...)
	if err != nil {
		panic(err)
	}
}

// RegisterRoutes register routes
func RegisterRoutes(e *gin.Engine) {
	e.NoRoute(handleNotFound)

	e.GET("/", handleHomePage)
	e.GET("/post/:slug", handleArticlePage)
	e.GET("/series.html", handleSeriesPage)
	e.GET("/archives.html", handleArchivePage)
	e.GET("/search.html", handleSearchPage)
	e.GET("/disqus/form/post-:slug", handleDisqusPage)
	e.GET("/beacon.html", handleBeaconPage)
}

// baseParams 基础参数
func baseParams(c *gin.Context) gin.H {
	version := 0

	cookie, err := c.Request.Cookie("v")
	if err != nil || cookie.Value !=
		fmt.Sprint(config.Conf.BlogApp.StaticVersion) {
		version = config.Conf.BlogApp.StaticVersion
	}
	return gin.H{
		"BlogName": cache.Ei.Blogger.BlogName,
		"SubTitle": cache.Ei.Blogger.SubTitle,
		"BTitle":   cache.Ei.Blogger.BTitle,
		"BeiAn":    cache.Ei.Blogger.BeiAn,
		"Domain":   config.Conf.BlogApp.Host,
		"CopyYear": time.Now().Year(),
		"Twitter":  config.Conf.BlogApp.Twitter,
		"Qiniu":    config.Conf.BlogApp.Qiniu,
		"Disqus":   config.Conf.BlogApp.Disqus,
		"Version":  version,
	}
}

// handleNotFound not found page
func handleNotFound(c *gin.Context) {
	params := baseParams(c)
	params["title"] = "Not Found"
	params["Description"] = "404 Not Found"
	params["Path"] = ""
	c.Status(http.StatusNotFound)
	renderHTMLHomeLayout(c, "notfound", params)
}

// handleHomePage 首页
func handleHomePage(c *gin.Context) {
	params := baseParams(c)
	params["title"] = cache.Ei.Blogger.BTitle + " | " + cache.Ei.Blogger.SubTitle
	params["Description"] = "博客首页，" + cache.Ei.Blogger.SubTitle
	params["Path"] = c.Request.URL.Path
	params["CurrentPage"] = "blog-home"
	pn, err := strconv.Atoi(c.Query("pn"))
	if err != nil || pn < 1 {
		pn = 1
	}
	params["Prev"], params["Next"], params["List"] = cache.Ei.PageArticles(pn,
		config.Conf.BlogApp.General.PageNum)

	renderHTMLHomeLayout(c, "home", params)
}

// handleArticlePage 文章页
func handleArticlePage(c *gin.Context) {
	slug := c.Param("slug")
	if !strings.HasSuffix(slug, ".html") || cache.Ei.ArticlesMap[slug[:len(slug)-5]] == nil {
		handleNotFound(c)
		return
	}
	article := cache.Ei.ArticlesMap[slug[:len(slug)-5]]
	params := baseParams(c)
	params["Title"] = article.Title + " | " + cache.Ei.Blogger.BTitle
	params["Path"] = c.Request.URL.Path
	params["CurrentPage"] = "post-" + article.Slug
	params["Article"] = article

	var name string
	switch slug {
	case "blogroll.html":
		name = "blogroll"
		params["Description"] = "友情连接，" + cache.Ei.Blogger.SubTitle
	case "about.html":
		name = "about"
		params["Description"] = "关于作者，" + cache.Ei.Blogger.SubTitle
	default:
		params["Description"] = article.Desc + "，" + cache.Ei.Blogger.SubTitle
		name = "article"
		params["Copyright"] = cache.Ei.Blogger.Copyright
		if !article.UpdateTime.IsZero() {
			params["Days"] = int(time.Now().Sub(article.UpdateTime).Hours()) / 24
		} else {
			params["Days"] = int(time.Now().Sub(article.CreateTime).Hours()) / 24
		}
		if article.SerieID > 0 {
			for _, series := range cache.Ei.Series {
				if series.ID == article.SerieID {
					params["Serie"] = series
				}
			}
		}
	}
	renderHTMLHomeLayout(c, name, params)
}

// handleSeriesPage 专题页
func handleSeriesPage(c *gin.Context) {
	params := baseParams(c)
	params["Title"] = "专题 | " + cache.Ei.Blogger.BTitle
	params["Description"] = "专题列表，" + cache.Ei.Blogger.SubTitle
	params["Path"] = c.Request.URL.Path
	params["CurrentPage"] = "series"
	params["Article"] = cache.Ei.PageSeries
	renderHTMLHomeLayout(c, "series", params)
}

// handleArchivePage 归档页
func handleArchivePage(c *gin.Context) {
	params := baseParams(c)
	params["Title"] = "归档 | " + cache.Ei.Blogger.BTitle
	params["Description"] = "博客归档，" + cache.Ei.Blogger.SubTitle
	params["Path"] = c.Request.URL.Path
	params["CurrentPage"] = "archives"
	params["Article"] = cache.Ei.PageArchives
	renderHTMLHomeLayout(c, "archives", params)
}

// handleSearchPage 搜索页
func handleSearchPage(c *gin.Context) {
	params := baseParams(c)
	params["Title"] = "站内搜索 | " + cache.Ei.Blogger.BTitle
	params["Description"] = "站内搜索，" + cache.Ei.Blogger.SubTitle
	params["Path"] = ""
	params["CurrentPage"] = "search-post"

	q := strings.TrimSpace(c.Query("q"))
	if q != "" {
		start, err := strconv.Atoi(c.Query("start"))
		if start < 1 || err != nil {
			start = 1
		}
		params["Word"] = q

		vals := c.Request.URL.Query()
		result, err := internal.ElasticSearch(q, config.Conf.BlogApp.General.PageNum, start-1)
		if err != nil {
			logrus.Error("HandleSearchPage.ElasticSearch: ", err)
		} else {
			result.Took /= 1000
			for i, v := range result.Hits.Hits {
				article := cache.Ei.ArticlesMap[v.Source.Slug]
				if len(v.Highlight.Content) == 0 && article != nil {
					result.Hits.Hits[i].Highlight.Content = []string{article.Excerpt}
				}
			}
			params["SearchResult"] = result
			if num := start - config.Conf.BlogApp.General.PageNum; num > 0 {
				vals.Set("start", fmt.Sprint(num))
				params["Prev"] = vals.Encode()
			}
			if num := start + config.Conf.BlogApp.General.PageNum; result.Hits.Total >= num {
				vals.Set("start", fmt.Sprint(num))
				params["Next"] = vals.Encode()
			}
		}
	} else {
		params["HotWords"] = config.Conf.BlogApp.HotWords
	}
	renderHTMLHomeLayout(c, "search", params)
}

// handleDisqusPage 评论页
func handleDisqusPage(c *gin.Context) {
	array := strings.Split(c.Param("slug"), "|")
	if len(array) != 4 || array[1] == "" {
		c.String(http.StatusOK, "出错啦。。。")
		return
	}
	article := cache.Ei.ArticlesMap[array[0]]
	params := gin.H{
		"Titile": "发表评论 | " + config.Conf.BlogApp.Blogger.BTitle,
		"ATitle": article.Title,
		"Thread": array[1],
		"Slug":   article.Slug,
	}
	err := htmlTmpl.ExecuteTemplate(c.Writer, "disqus.html", params)
	if err != nil {
		panic(err)
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
}

// handleBeaconPage 服务端推送谷歌统计
func handleBeaconPage(c *gin.Context) {
	ua := c.Request.UserAgent()

	vals := c.Request.URL.Query()
	vals.Set("v", config.Conf.BlogApp.Google.V)
	vals.Set("tid", config.Conf.BlogApp.Google.Tid)
	vals.Set("t", config.Conf.BlogApp.Google.T)
	cookie, _ := c.Cookie("u")
	vals.Set("cid", cookie)

	vals.Set("dl", c.Request.Referer())
	vals.Set("uip", c.ClientIP())
	go func() {
		req, err := http.NewRequest("POST", config.Conf.BlogApp.Google.URL,
			strings.NewReader(vals.Encode()))
		if err != nil {
			logrus.Error("HandleBeaconPage.NewRequest: ", err)
			return
		}
		req.Header.Set("User-Agent", ua)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			logrus.Error("HandleBeaconPage.Do: ", err)
			return
		}
		defer res.Body.Close()
		data, err := ioutil.ReadAll(res.Body)
		if err != nil {
			logrus.Error("HandleBeaconPage.ReadAll: ", err)
			return
		}
		if res.StatusCode/100 != 2 {
			logrus.Error(string(data))
		}
	}()
	c.Status(http.StatusNoContent)
}

// renderHTMLHomeLayout homelayout html
func renderHTMLHomeLayout(c *gin.Context, name string, data gin.H) {
	buf := bytes.Buffer{}
	err := htmlTmpl.ExecuteTemplate(&buf, name, data)
	if err != nil {
		panic(err)
	}
	data["LayoutContent"] = htemplate.HTML(buf.String())
	err = htmlTmpl.ExecuteTemplate(c.Writer, "homelayout.html", data)
	if err != nil {
		panic(err)
	}
	if c.Writer.Status() == 0 {
		c.Status(http.StatusOK)
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
}
