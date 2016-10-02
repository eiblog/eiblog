// Package main provides ...
package main

import (
	"fmt"
	"html/template"
	"time"

	"github.com/eiblog/eiblog/setting"
	"github.com/eiblog/utils/logd"
	"github.com/eiblog/utils/tmpl"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

var router *gin.Engine

func init() {
	if setting.Conf.RunMode == setting.PROD {
		gin.SetMode(gin.ReleaseMode)
		// set log print level
		logd.SetLevel(logd.Lerror)
	}
	router = gin.Default()
	store := sessions.NewCookieStore([]byte("eiblog321"))
	store.Options(sessions.Options{
		MaxAge:   86400 * 999,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
	})
	router.Use(sessions.Sessions("su", store))
	// 匹配模版
	//router.LoadHTMLGlob("views/*.html")
	if tmpl, err := template.New("").Funcs(tmpl.TplFuncMap).ParseGlob("views/*.*"); err == nil {
		tmpl, err = tmpl.ParseGlob("views/admin/*.html")
		if err != nil {
			logd.Fatal(err)
		}
		router.SetHTMLTemplate(tmpl)
	} else {
		logd.Fatal(err)
	}
	// 开启静态文件
	router.Static("/static", "./static")
	router.Use(Filter())
	router.NoRoute(HandleNotFound)
	router.GET("/", HandleHomePage)
	router.GET("/post/:slug", HandleArticlePage)
	router.GET("/series.html", HandleSeriesPage)
	router.GET("/archives.html", HandleArchivesPage)
	router.GET("/search.html", HandleSearchPage)
	router.GET("/beacon.html", HandleBeacon)
	router.GET("/disqus/:slug", HandleDisqus)
	router.GET("/rss.html", HandleFeed)
	router.GET("/feed", HandleFeed)
	router.GET("/opensearch.xml", HandleOpenSearch)
	router.GET("/sitemap.xml", HandleSitemap)
	router.GET("/robots.txt", HandleRobots)
	// 后台相关
	admin := router.Group("/admin")
	admin.GET("/login", HandleLogin)
	admin.POST("/login", HandleLoginPost)
	auth := admin.Use(AuthFilter())
	{
		// console
		auth.GET("/profile", HandleProfile)
		auth.GET("/plugins", HandlePlugins)
		auth.GET("/themes", HandleThemes)
		// write
		auth.GET("/write-post", HandlePost)
		// manage
		auth.GET("/manage-posts", HandlePosts)
		auth.GET("/manage-series", HandleSeries)
		auth.GET("/add-serie", HandleSerie)
		auth.GET("/manage-tags", HandleTags)
		auth.GET("/manage-draft", HandleDraft)
		auth.GET("/manage-trash", HandleTrash)
		auth.GET("/options-general", HandleGeneral)
		auth.GET("/options-discussion", HandleDiscussion)
		auth.GET("/draft-delete", HandleDraftDelete)
		// api
		auth.POST("/api/:action", HandleAPI)
	}
}

func Run() {
	var (
		endRunning = make(chan bool, 1)
		err        error
	)
	if setting.Conf.Mode.EnableHttp {
		go func() {
			logd.Info(fmt.Sprintf("http server Running on %d", setting.Conf.Mode.HttpPort))
			err = router.Run(fmt.Sprintf(":%d", setting.Conf.Mode.HttpPort))
			if err != nil {
				logd.Info("ListenAndServe: ", err)
				time.Sleep(100 * time.Microsecond)
				endRunning <- true
			}
		}()
	}
	if setting.Conf.Mode.EnableHttps {
		go func() {
			logd.Info(fmt.Sprintf("https server Running on %d", setting.Conf.Mode.HttpsPort))
			err = router.RunTLS(fmt.Sprintf(":%d", setting.Conf.Mode.HttpsPort), setting.Conf.Mode.CertFile, setting.Conf.Mode.KeyFile)
			if err != nil {
				logd.Info("ListenAndServe: ", err)
				time.Sleep(100 * time.Microsecond)
				endRunning <- true
			}
		}()
	}
	<-endRunning
}
