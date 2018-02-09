// Package main provides ...
package main

import (
	"crypto/rand"
	"fmt"
	"text/template"
	"time"

	"github.com/eiblog/eiblog/setting"
	"github.com/eiblog/utils/logd"
	"github.com/eiblog/utils/tmpl"
	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

var (
	router *gin.Engine
	Tmpl   *template.Template
)

func init() {
	// 运行模式
	if setting.Conf.RunMode == setting.PROD {
		gin.SetMode(gin.ReleaseMode)
		logd.SetLevel(logd.Lerror)
	}

	router = gin.Default()
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		logd.Fatal(err)
	}
	store := sessions.NewCookieStore(b)
	store.Options(sessions.Options{
		MaxAge:   86400 * 7,
		Path:     "/",
		Secure:   setting.Conf.RunMode == setting.PROD,
		HttpOnly: true,
	})
	router.Use(sessions.Sessions("su", store))
	// 匹配模版
	Tmpl = template.New("eiblog").Funcs(tmpl.TplFuncMap)
	files := ReadDir("views", func(name string) bool {
		if name == ".DS_Store" {
			return true
		}
		return false
	})
	_, err = Tmpl.ParseFiles(files...)
	if err != nil {
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
	router.GET("/disqus/post-:slug", HandleDisqus)
	router.GET("/disqus/form/post-:slug", HandleDisqusFrom)
	router.POST("/disqus/create", HandleDisqusCreate)
	router.GET("/rss.html", HandleFeed)
	router.GET("/feed", HandleFeed)
	router.GET("/opensearch.xml", HandleOpenSearch)
	router.GET("/sitemap.xml", HandleSitemap)
	router.GET("/robots.txt", HandleRobots)
	router.GET("/crossdomain.xml", HandleCrossDomain)
	router.GET("/favicon.ico", HandleFavicon)
	// 后台相关
	admin := router.Group("/admin")
	admin.GET("/login", HandleLogin)
	admin.POST("/login", HandleLoginPost)
	auth := admin.Use(AuthFilter())
	{
		// console
		auth.GET("/profile", HandleProfile)
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

// 开始运行
func Run() {
	var (
		endRunning = make(chan bool, 1)
		err        error
	)
	if setting.Conf.Mode.EnableHttp {
		go func() {
			logd.Printf("http server Running on %d\n", setting.Conf.Mode.HttpPort)
			err = router.Run(fmt.Sprintf(":%d", setting.Conf.Mode.HttpPort))
			if err != nil {
				logd.Error("ListenAndServe: ", err)
				time.Sleep(100 * time.Microsecond)
				endRunning <- true
			}
		}()
	}
	if setting.Conf.Mode.EnableHttps {
		if setting.Conf.Mode.AutoCert {
			go func() {
				logd.Print("https server Running on 443")
				err = autotls.Run(router, setting.Conf.Mode.Domain)
				if err != nil {
					logd.Error("ListenAndServe: ", err)
					time.Sleep(100 * time.Microsecond)
					endRunning <- true
				}
			}()
		} else {
			go func() {
				logd.Printf("https server Running on %d\n", setting.Conf.Mode.HttpsPort)
				err = router.RunTLS(fmt.Sprintf(":%d", setting.Conf.Mode.HttpsPort),
					setting.Conf.Mode.CertFile, setting.Conf.Mode.KeyFile)
				if err != nil {
					logd.Error("ListenAndServe: ", err)
					time.Sleep(100 * time.Microsecond)
					endRunning <- true
				}
			}()
		}
	}
	<-endRunning
}
