package internal

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/eiblog/eiblog/cmd/eiblog/config"

	"github.com/sirupsen/logrus"
)

func startTimer() {
	err := generateOpensearch()
	if err != nil {
		logrus.Error("startTimer.generateOpensearch: ", err)
	}
	err = generateRobots()
	if err != nil {
		logrus.Error("startTimer.generateRobots: ", err)
	}
	err = generateCrossdomain()
	if err != nil {
		logrus.Error("startTimer.generateCrossdomain: ", err)
	}
	// 定时刷新
	refreshFeedAndSitemap()
}

// generateOpensearch 生成opensearch.xml
func generateOpensearch() error {
	tpl := XMLTemplate.Lookup("opensearchTpl.xml")
	if tpl == nil {
		return errors.New("not found: opensearchTpl.xml")
	}

	params := map[string]string{
		"BTitle":   Ei.Blogger.BTitle,
		"SubTitle": Ei.Blogger.SubTitle,
		"Host":     config.Conf.Host,
	}
	path := filepath.Join(config.EtcDir, "assets", "opensearch.xml")
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	return tpl.Execute(f, params)
}

// generateRobots 生成robots.txt
func generateRobots() error {
	tpl := XMLTemplate.Lookup("robotsTpl.xml")
	if tpl == nil {
		return errors.New("not found: robotsTpl.xml")
	}

	params := map[string]string{
		"Host": config.Conf.Host,
	}
	path := filepath.Join(config.EtcDir, "assets", "robots.txt")
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	return tpl.Execute(f, params)
}

// generateCrossdomain 生成crossdomain.xml
func generateCrossdomain() error {
	tpl := XMLTemplate.Lookup("crossdomainTpl.xml")
	if tpl == nil {
		return errors.New("not found: crossdomainTpl.xml")
	}

	params := map[string]string{
		"Host": config.Conf.Host,
	}
	path := filepath.Join(config.EtcDir, "assets", "crossdomain.xml")
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	return tpl.Execute(f, params)
}

// refreshFeedAndSitemap 定时刷新feed和sitemap
func refreshFeedAndSitemap() {
	defer time.AfterFunc(time.Hour*4, refreshFeedAndSitemap)

	now := time.Now()
	// generate feed & sitemap
	err := generateFeed()
	if err != nil {
		logrus.Error("startTimer.generateFeed: ", err)
	}
	err = generateSitemap()
	if err != nil {
		logrus.Error("startTimer.generateSitemap: ", err)
	}
	// clean expired articles
	exp := now.Add(-48 * time.Hour)
	err = Store.CleanArticles(context.Background(), exp)
	if err != nil {
		logrus.Error("startTimer.CleanArticles: ", err)
	}
	// fetch disqus count
	err = DisqusClient.PostsCount(Ei.ArticlesMap)
	if err != nil {
		logrus.Error("startTimer.PostsCount: ", err)
	}
}

// generateFeed 定时刷新feed
func generateFeed() error {
	tpl := XMLTemplate.Lookup("feedTpl.xml")
	if tpl == nil {
		return errors.New("not found: feedTpl.xml")
	}

	_, _, articles := Ei.PageArticleFE(1, 20)
	params := map[string]interface{}{
		"Title":     Ei.Blogger.BTitle,
		"SubTitle":  Ei.Blogger.SubTitle,
		"Host":      config.Conf.Host,
		"FeedrURL":  config.Conf.FeedRPC.FeedrURL,
		"BuildDate": time.Now().Format(time.RFC1123Z),
		"Articles":  articles,
	}
	path := filepath.Join(config.EtcDir, "assets", "feed.xml")
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	return tpl.Execute(f, params)
}

// generateSitemap 定时刷新sitemap
func generateSitemap() error {
	tpl := XMLTemplate.Lookup("sitemapTpl.xml")
	if tpl == nil {
		return errors.New("not found: sitemapTpl.xml")
	}

	params := map[string]interface{}{
		"Articles": Ei.Articles,
		"Host":     config.Conf.Host,
	}
	path := filepath.Join(config.EtcDir, "assets", "sitemap.xml")
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	return tpl.Execute(f, params)
}
