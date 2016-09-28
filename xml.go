// Package main provides ...
// generate feed.xml and sitemap.xml
package main

import (
	"os"
	"text/template"
	"time"

	"github.com/EiBlog/utils/logd"
	"github.com/EiBlog/utils/tmpl"
)

const (
	FEED_COUNT    = 20
	TEMPLATE_GLOB = "conf/tpl/*.xml"
)

var tpls *template.Template

func init() {
	var err error
	tpls, err = template.New("").Funcs(template.FuncMap{"dateformat": tmpl.DateFormat}).ParseGlob(TEMPLATE_GLOB)
	if err != nil {
		logd.Fatal(err)
	}

	go doFeed()
	go doSitemap()
}

func doFeed() {
	_, _, artcs := PageList(1, FEED_COUNT)
	buildDate := time.Now()
	params := map[string]interface{}{
		"Title":       Ei.BTitle,
		"SubTitle":    Ei.SubTitle,
		"Domain":      runmode.Domain,
		"Enablehttps": runmode.EnableHttps,
		"BuildDate":   buildDate.Format(time.RFC1123Z),
		"Artcs":       artcs,
	}

	f, err := os.OpenFile("conf/feed.xml", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		logd.Error(err)
		return
	}
	defer f.Close()
	tpl := tpls.Lookup("feedTpl.xml")
	if tpl == nil {
		logd.Error(err)
		return
	}
	err = tpl.Execute(f, params)
	if err != nil {
		logd.Error(err)
		return
	}
	time.AfterFunc(time.Hour*4, doFeed)
}

func doSitemap() {
	params := map[string]interface{}{"Artcs": Ei.Articles, "Domain": runmode.Domain, "Enablehttps": runmode.EnableHttps}
	f, err := os.OpenFile("conf/sitemap.xml", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		logd.Error(err)
		return
	}
	defer f.Close()
	tpl := tpls.Lookup("sitemapTpl.xml")
	if tpl == nil {
		logd.Error(err)
		return
	}
	err = tpl.Execute(f, params)
	if err != nil {
		logd.Error(err)
		return
	}
	time.AfterFunc(time.Hour*24, doFeed)
}
