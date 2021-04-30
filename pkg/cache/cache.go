// Package cache provides ...
package cache

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/eiblog/eiblog/pkg/cache/render"
	"github.com/eiblog/eiblog/pkg/cache/store"
	"github.com/eiblog/eiblog/pkg/config"
	"github.com/eiblog/eiblog/pkg/internal"
	"github.com/eiblog/eiblog/pkg/model"
	"github.com/eiblog/eiblog/tools"

	"github.com/sirupsen/logrus"
)

var (
	// Ei eiblog cache
	Ei *Cache

	// regenerate pages chan
	PagesCh     = make(chan string, 2)
	PageSeries  = "series-md"
	PageArchive = "archive-md"
)

func init() {
	// init timezone
	var err error
	tools.TimeLocation, err = time.LoadLocation(
		config.Conf.EiBlogApp.General.Timezone)
	if err != nil {
		panic(err)
	}
	// init store
	store, err := store.NewStore(config.Conf.Database.Driver,
		config.Conf.Database.Source)
	if err != nil {
		panic(err)
	}
	// Ei init
	Ei = &Cache{
		lock:        sync.Mutex{},
		Store:       store,
		TagArticles: make(map[string]model.SortedArticles),
		ArticlesMap: make(map[string]*model.Article),
	}
	err = Ei.loadOrInit()
	if err != nil {
		panic(err)
	}
	go Ei.regeneratePages()
	go Ei.timerClean()
	go Ei.timerDisqus()
}

// Cache 整站缓存
type Cache struct {
	lock sync.Mutex
	store.Store

	// load from db
	Blogger  *model.Blogger
	Account  *model.Account
	Articles model.SortedArticles

	// auto generate
	PageSeries   string // page
	Series       model.SortedSeries
	PageArchives string // page
	Archives     model.SortedArchives
	TagArticles  map[string]model.SortedArticles // tagname:articles
	ArticlesMap  map[string]*model.Article       // slug:article
}

// AddArticle 添加文章
func (c *Cache) AddArticle(article *model.Article) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	// store
	err := c.InsertArticle(context.Background(), article,
		config.Conf.EiBlogApp.General.StartID)
	if err != nil {
		return err
	}
	// 是否是草稿
	if article.IsDraft {
		return nil
	}
	// 正式发布文章
	c.refreshCache(article, false)
	return nil
}

// RepArticle 替换文章
func (c *Cache) RepArticle(oldArticle, newArticle *model.Article) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.ArticlesMap[newArticle.Slug] = newArticle
	render.GenerateExcerptMarkdown(newArticle)
	if newArticle.ID < config.Conf.EiBlogApp.General.StartID {
		return
	}
	if oldArticle != nil { // 移除旧文章
		c.refreshCache(oldArticle, true)
	}
	c.refreshCache(newArticle, false)
}

// DelArticle 删除文章
func (c *Cache) DelArticle(id int) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	article, _ := c.FindArticleByID(id)
	if article == nil {
		return nil
	}
	// set delete
	err := c.UpdateArticle(context.Background(), id, map[string]interface{}{
		"deleted_at": time.Now(),
	})
	if err != nil {
		return err
	}
	// drop from tags,series,archives
	c.refreshCache(article, true)
	return nil
}

// AddSerie 添加专题
func (c *Cache) AddSerie(serie *model.Serie) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	err := c.InsertSerie(context.Background(), serie)
	if err != nil {
		return err
	}
	c.Series = append(c.Series, serie)
	PagesCh <- PageSeries
	return nil
}

// DelSerie 删除专题
func (c *Cache) DelSerie(id int) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	for i, serie := range c.Series {
		if serie.ID == id {
			if len(serie.Articles) > 0 {
				return errors.New("请删除该专题下的所有文章")
			}
			err := c.RemoveSerie(context.Background(), id)
			if err != nil {
				return err
			}
			c.Series[i] = nil
			c.Series = append(c.Series[:i], c.Series[i+1:]...)
			PagesCh <- PageSeries
			break
		}
	}
	return nil
}

// PageArticleFE 文章翻页
func (c *Cache) PageArticleFE(page int, pageSize int) (prev,
	next int, articles []*model.Article) {

	var l int
	for l = len(c.Articles); l > 0; l-- {
		if c.Articles[l-1].ID >= config.Conf.EiBlogApp.General.StartID {
			break
		}
	}
	if l == 0 {
		return 0, 0, nil
	}
	m := l / pageSize
	if d := l % pageSize; d > 0 {
		m++
	}
	if page > m {
		page = m
	}
	if page > 1 {
		prev = page - 1
	}
	if page < m {
		next = page + 1
	}
	s := (page - 1) * pageSize
	e := page * pageSize
	if e > l {
		e = l
	}
	articles = c.Articles[s:e]
	return
}

// PageArticleBE 后台文章分页
func (c *Cache) PageArticleBE(se int, kw string, draft, del bool, p,
	n int) ([]*model.Article, int) {

	search := store.SearchArticles{
		Page:   p,
		Limit:  n,
		Fields: make(map[string]interface{}),
	}
	if draft {
		search.Fields[store.SearchArticleDraft] = true
	} else if del {
		search.Fields[store.SearchArticleTrash] = true
	} else {
		search.Fields[store.SearchArticleDraft] = false
		if se > 0 {
			search.Fields[store.SearchArticleSerieID] = se
		}
		if kw != "" {
			search.Fields[store.SearchArticleTitle] = kw
		}
	}
	articles, count, err := c.LoadArticleList(context.Background(), search)
	if err != nil {
		return nil, 0
	}
	max := count / n
	if count%n > 0 {
		max++
	}
	return articles, max
}

// FindArticleByID 通过ID查找文章
func (c *Cache) FindArticleByID(id int) (*model.Article, int) {
	for i, article := range c.Articles {
		if article.ID == id {
			return article, i
		}
	}
	return nil, -1
}

// refreshCache 刷新缓存
func (c *Cache) refreshCache(article *model.Article, del bool) {
	if del {
		_, idx := c.FindArticleByID(article.ID)

		delete(c.ArticlesMap, article.Slug)
		c.Articles = append(c.Articles[:idx], c.Articles[idx+1:]...)
		// 从链表移除
		c.recalcLinkedList(article, true)
		// 从tag、serie、archive移除
		c.redelArticle(article)
		return
	}
	// 添加文章
	defer render.GenerateExcerptMarkdown(article)

	c.ArticlesMap[article.Slug] = article
	c.Articles = append([]*model.Article{article}, c.Articles...)
	sort.Sort(c.Articles)
	// 从链表添加
	c.recalcLinkedList(article, false)
	// 从tag、serie、archive添加
	c.readdArticle(article, true)
}

// recalcLinkedList 重算文章链表
func (c *Cache) recalcLinkedList(article *model.Article, del bool) {
	// 删除操作
	if del {
		if article.Prev == nil && article.Next != nil {
			article.Next.Prev = nil
		} else if article.Prev != nil && article.Next == nil {
			article.Prev.Next = nil
		} else if article.Prev != nil && article.Next != nil {
			article.Prev.Next = article.Next
			article.Next.Prev = article.Prev
		}
		return
	}
	// 添加操作
	_, idx := c.FindArticleByID(article.ID)
	if idx == 0 && c.Articles[idx+1].ID >=
		config.Conf.EiBlogApp.General.StartID {
		article.Next = c.Articles[idx+1]
		c.Articles[idx+1].Prev = article
	} else if idx > 0 && c.Articles[idx-1].ID >=
		config.Conf.EiBlogApp.General.StartID {
		article.Prev = c.Articles[idx-1]
		if c.Articles[idx-1].Next != nil {
			article.Next = c.Articles[idx-1].Next
			c.Articles[idx-1].Next.Prev = article
		}
		c.Articles[idx-1].Next = article
	}
}

// readdArticle 添加文章到tag、series、archive
func (c *Cache) readdArticle(article *model.Article, needSort bool) {
	// tag
	for _, tag := range article.Tags {
		c.TagArticles[tag] = append(c.TagArticles[tag], article)
		if needSort {
			sort.Sort(c.TagArticles[tag])
		}
	}
	// series
	for i, serie := range c.Series {
		if serie.ID == article.SerieID {
			c.Series[i].Articles = append(c.Series[i].Articles, article)
			if needSort {
				sort.Sort(c.Series[i].Articles)
				PagesCh <- PageSeries // 重建专题
			}
		}
	}
	// archive
	y, m, _ := article.CreatedAt.Date()
	for i, archive := range c.Archives {
		if ay, am, _ := archive.Time.Date(); y == ay && m == am {
			c.Archives[i].Articles = append(c.Archives[i].Articles, article)
		}
		if needSort {
			sort.Sort(c.Archives[i].Articles)
			PagesCh <- PageArchive // 重建归档
		}
		return
	}
	// 新建归档
	c.Archives = append(c.Archives, &model.Archive{
		Time:     article.CreatedAt,
		Articles: model.SortedArticles{article},
	})
	if needSort { // 重建归档
		PagesCh <- PageArchive
	}
}

// redelArticle 从tag、series、archive删除文章
func (c *Cache) redelArticle(article *model.Article) {
	// tag
	for _, tag := range article.Tags {
		for i, v := range c.TagArticles[tag] {
			if v == article {
				c.TagArticles[tag] = append(c.TagArticles[tag][0:i], c.TagArticles[tag][i+1:]...)
				if len(c.TagArticles[tag]) == 0 {
					delete(c.TagArticles, tag)
				}
			}
		}
	}
	// serie
	for i, serie := range c.Series {
		if serie.ID == article.SerieID {
			for j, v := range serie.Articles {
				if v == article {
					c.Series[i].Articles = append(c.Series[i].Articles[0:j],
						c.Series[i].Articles[j+1:]...)
					PagesCh <- PageSeries
					break
				}
			}
		}
	}
	// archive
	for i, archive := range c.Archives {
		ay, am, _ := archive.Time.Date()
		if y, m, _ := article.CreatedAt.Date(); ay == y && am == m {
			for j, v := range archive.Articles {
				if v == article {
					c.Archives[i].Articles = append(c.Archives[i].Articles[0:j],
						c.Archives[i].Articles[j+1:]...)
					if len(c.Archives[i].Articles) == 0 {
						c.Archives = append(c.Archives[:i], c.Archives[i+1:]...)
					}
					PagesCh <- PageArchive
					break
				}
			}
		}
	}
}

// loadOrInit 读取数据或初始化
func (c *Cache) loadOrInit() error {
	blogapp := config.Conf.EiBlogApp
	// blogger
	blogger := &model.Blogger{
		BlogName:  strings.Title(blogapp.Account.Username),
		SubTitle:  "Rome was not built in one day.",
		BeiAn:     "蜀ICP备xxxxxxxx号-1",
		BTitle:    fmt.Sprintf("%s's Blog", strings.Title(blogapp.Account.Username)),
		Copyright: `本站使用「<a href="//creativecommons.org/licenses/by/4.0/">署名 4.0 国际</a>」创作共享协议，转载请注明作者及原网址。`,
	}
	created, err := c.LoadInsertBlogger(context.Background(), blogger)
	if err != nil {
		return err
	}
	c.Blogger = blogger
	if created { // init articles: about blogroll
		about := &model.Article{
			ID:        1, // 固定ID
			Author:    blogapp.Account.Username,
			Title:     "关于",
			Slug:      "about",
			CreatedAt: time.Time{},
		}
		err = c.InsertArticle(context.Background(), about,
			config.Conf.EiBlogApp.General.StartID)
		if err != nil {
			return err
		}
		// 推送到 disqus
		go internal.ThreadCreate(about, blogger.BTitle)
		blogroll := &model.Article{
			ID:        2, // 固定ID
			Author:    blogapp.Account.Username,
			Title:     "友情链接",
			Slug:      "blogroll",
			CreatedAt: time.Time{}.AddDate(0, 0, 7),
		}
		err = c.InsertArticle(context.Background(), blogroll,
			config.Conf.EiBlogApp.General.StartID)
		if err != nil {
			return err
		}
	}
	// account
	pwd := tools.EncryptPasswd(blogapp.Account.Password,
		blogapp.Account.Password)

	account := &model.Account{
		Username: blogapp.Account.Username,
		Password: pwd,
	}
	_, err = c.LoadInsertAccount(context.Background(), account)
	if err != nil {
		return err
	}
	c.Account = account
	// series
	series, err := c.LoadAllSerie(context.Background())
	if err != nil {
		return err
	}
	c.Series = series
	// all articles
	search := store.SearchArticles{
		Page:   1,
		Limit:  9999,
		Fields: map[string]interface{}{store.SearchArticleDraft: false},
	}
	articles, _, err := c.LoadArticleList(context.Background(), search)
	if err != nil {
		return err
	}
	for i, v := range articles {
		// 渲染页面
		render.GenerateExcerptMarkdown(v)

		c.ArticlesMap[v.Slug] = v
		// 分析文章
		if v.ID < blogapp.General.StartID {
			continue
		}
		if i > 0 {
			v.Prev = articles[i-1]
		}
		if i < len(articles)-1 &&
			articles[i+1].ID >= blogapp.General.StartID {
			v.Next = articles[i+1]
		}
		c.readdArticle(v, false)
	}
	Ei.Articles = articles
	// 重建专题与归档
	PagesCh <- PageSeries
	PagesCh <- PageArchive
	return nil
}

// regeneratePages 重新生成series,archive页面
func (c *Cache) regeneratePages() {
	for {
		switch page := <-PagesCh; page {
		case PageSeries:
			sort.Sort(c.Series)
			buf := bytes.Buffer{}
			buf.WriteString(c.Blogger.SeriesSay)
			buf.WriteString("\n\n")
			for _, series := range c.Series {
				buf.WriteString(fmt.Sprintf("### %s{#toc-%d}", series.Name, series.ID))
				buf.WriteByte('\n')
				buf.WriteString(series.Desc)
				buf.WriteString("\n\n")
				for _, article := range series.Articles {
					//eg. * [标题一](/post/hello-world.html) <span class="date">(Man 02, 2006)</span>
					str := fmt.Sprintf(`* [%s](/post/%s.html) <span class="date">(%s)</span>`,
						article.Title, article.Slug, article.CreatedAt.Format("Jan 02, 2006"))
					buf.WriteString(str)
				}
				buf.WriteString("\n\n")
			}
			c.PageSeries = string(render.RenderPage(buf.Bytes()))
		case PageArchive:
			sort.Sort(c.Archives)
			buf := bytes.Buffer{}
			buf.WriteString(c.Blogger.ArchivesSay + "\n")
			var (
				currentYear string
				gt12Month   = len(c.Archives) > 12
			)
			for _, archive := range c.Archives {
				t := archive.Time.In(tools.TimeLocation)
				if gt12Month {
					year := t.Format("2006 年")
					if currentYear != year {
						currentYear = year
						buf.WriteString(fmt.Sprintf("\n### %s\n\n", t.Format("2006 年")))
					}
				} else {
					buf.WriteString(fmt.Sprintf("\n### %s\n\n", t.Format("2006年1月")))
				}
				for i, article := range archive.Articles {
					createdAt := article.CreatedAt.In(tools.TimeLocation)
					if i == 0 && gt12Month {
						str := fmt.Sprintf(`* *[%s](/post/%s.html) <span class="date">(%s)</span>`,
							article.Title, article.Slug, createdAt.Format("Jan 02, 2006"))
						buf.WriteString(str)
					} else {
						str := fmt.Sprintf(`* [%s](/post/%s.html) <span class="date">(%s)</span>`,
							article.Title, article.Slug, createdAt.Format("Jan 02, 2006"))
						buf.WriteString(str)
					}
					buf.WriteByte('\n')
				}
			}
			c.PageArchives = string(render.RenderPage(buf.Bytes()))
		}
	}
}

// timerClean 定时清理文章
func (c *Cache) timerClean() {
	ticker := time.NewTicker(time.Hour)

	for range ticker.C {
		err := c.CleanArticles(context.Background())
		if err != nil {
			logrus.Error("cache.timerClean.CleanArticles: ", err)
		}
	}
}

// timerDisqus disqus定时操作
func (c *Cache) timerDisqus() {
	ticker := time.NewTicker(5 * time.Hour)

	for range ticker.C {
		err := internal.PostsCount(c.ArticlesMap)
		if err != nil {
			logrus.Error("cache.timerDisqus.PostsCount: ", err)
		}
	}
}
