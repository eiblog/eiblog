// Package internal provides ...
package internal

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/eiblog/eiblog/cmd/eiblog/config"
	"github.com/eiblog/eiblog/cmd/eiblog/handler/internal/store"
	"github.com/eiblog/eiblog/pkg/model"
	"github.com/eiblog/eiblog/tools"
)

var (
	// PagesCh regenerate pages chan
	PagesCh = make(chan string, 2)
	// PageSeries the page series regenerate flag
	PageSeries = "series-md"
	// PageArchive the page archive regenerate flag
	PageArchive = "archive-md"

	// ArticleStartID article start id
	ArticleStartID = 11
)

// Cache 整站缓存
type Cache struct {
	lock sync.Mutex

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

// NewCache 缓存整个博客数据
func NewCache() (*Cache, error) {
	// Ei init
	cache := &Cache{
		lock:        sync.Mutex{},
		TagArticles: make(map[string]model.SortedArticles),
		ArticlesMap: make(map[string]*model.Article),
	}
	err := cache.loadOrInit()
	if err != nil {
		return nil, err
	}
	// 异步渲染series,archive页面
	go cache.regeneratePages()
	return cache, nil
}

// AddArticle 添加文章
func (cache *Cache) AddArticle(article *model.Article) error {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	// store
	err := Store.InsertArticle(context.Background(), article, ArticleStartID)
	if err != nil {
		return err
	}
	// 是否是草稿
	if article.IsDraft {
		return nil
	}
	// 正式发布文章
	cache.refreshCache(article, false)
	return nil
}

// RepArticle 替换文章
func (cache *Cache) RepArticle(oldArticle, newArticle *model.Article) {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	cache.ArticlesMap[newArticle.Slug] = newArticle
	GenerateExcerptMarkdown(newArticle)
	if newArticle.ID < ArticleStartID {
		return
	}
	if oldArticle != nil { // 移除旧文章
		cache.refreshCache(oldArticle, true)
	}
	cache.refreshCache(newArticle, false)
}

// DelArticle 删除文章
func (cache *Cache) DelArticle(id int) error {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	article, _ := cache.FindArticleByID(id)
	if article == nil {
		return nil
	}
	// set delete
	err := Store.UpdateArticle(context.Background(), id, map[string]interface{}{
		"deleted_at": time.Now(),
	})
	if err != nil {
		return err
	}
	// drop from tags,series,archives
	cache.refreshCache(article, true)
	return nil
}

// AddSerie 添加专题
func (cache *Cache) AddSerie(serie *model.Serie) error {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	err := Store.InsertSerie(context.Background(), serie)
	if err != nil {
		return err
	}
	cache.Series = append(cache.Series, serie)
	PagesCh <- PageSeries
	return nil
}

// DelSerie 删除专题
func (cache *Cache) DelSerie(id int) error {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	for i, serie := range cache.Series {
		if serie.ID == id {
			if len(serie.Articles) > 0 {
				return errors.New("请删除该专题下的所有文章")
			}
			err := Store.RemoveSerie(context.Background(), id)
			if err != nil {
				return err
			}
			cache.Series[i] = nil
			cache.Series = append(cache.Series[:i], cache.Series[i+1:]...)
			PagesCh <- PageSeries
			break
		}
	}
	return nil
}

// PageArticleFE 文章翻页
func (cache *Cache) PageArticleFE(page int, pageSize int) (prev,
	next int, articles []*model.Article) {

	var l int
	for l = len(cache.Articles); l > 0; l-- {
		if cache.Articles[l-1].ID >= ArticleStartID {
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
	articles = cache.Articles[s:e]
	return
}

// PageArticleBE 后台文章分页
func (cache *Cache) PageArticleBE(se int, kw string, draft, del bool, p,
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
	articles, count, err := Store.LoadArticleList(context.Background(), search)
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
func (cache *Cache) FindArticleByID(id int) (*model.Article, int) {
	for i, article := range cache.Articles {
		if article.ID == id {
			return article, i
		}
	}
	return nil, -1
}

// refreshCache 刷新缓存
func (cache *Cache) refreshCache(article *model.Article, del bool) {
	if del {
		_, idx := cache.FindArticleByID(article.ID)

		delete(cache.ArticlesMap, article.Slug)
		cache.Articles = append(cache.Articles[:idx], cache.Articles[idx+1:]...)
		// 从链表移除
		cache.recalcLinkedList(article, true)
		// 从tag、serie、archive移除
		cache.redelArticle(article)
		return
	}
	// 添加文章
	defer GenerateExcerptMarkdown(article)

	cache.ArticlesMap[article.Slug] = article
	cache.Articles = append([]*model.Article{article}, cache.Articles...)
	sort.Sort(cache.Articles)
	// 从链表添加
	cache.recalcLinkedList(article, false)
	// 从tag、serie、archive添加
	cache.readdArticle(article, true)
}

// recalcLinkedList 重算文章链表
func (cache *Cache) recalcLinkedList(article *model.Article, del bool) {
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
	_, idx := cache.FindArticleByID(article.ID)
	if idx == 0 && cache.Articles[idx+1].ID >= ArticleStartID {
		article.Next = cache.Articles[idx+1]
		cache.Articles[idx+1].Prev = article
	} else if idx > 0 && cache.Articles[idx-1].ID >= ArticleStartID {
		article.Prev = cache.Articles[idx-1]
		if cache.Articles[idx-1].Next != nil {
			article.Next = cache.Articles[idx-1].Next
			cache.Articles[idx-1].Next.Prev = article
		}
		cache.Articles[idx-1].Next = article
	}
}

// readdArticle 添加文章到tag、series、archive
func (cache *Cache) readdArticle(article *model.Article, needSort bool) {
	// tag
	for _, tag := range article.Tags {
		cache.TagArticles[tag] = append(cache.TagArticles[tag], article)
		if needSort {
			sort.Sort(cache.TagArticles[tag])
		}
	}
	// series
	for i, serie := range cache.Series {
		if serie.ID != article.SerieID {
			continue
		}
		cache.Series[i].Articles = append(cache.Series[i].Articles, article)
		if needSort {
			sort.Sort(cache.Series[i].Articles)
			PagesCh <- PageSeries // 重建专题
		}
	}
	// archive
	y, m, _ := article.CreatedAt.Date()
	for i, archive := range cache.Archives {
		ay, am, _ := archive.Time.Date()
		if y != ay || m != am {
			continue
		}
		cache.Archives[i].Articles = append(cache.Archives[i].Articles, article)
		if needSort {
			sort.Sort(cache.Archives[i].Articles)
			PagesCh <- PageArchive // 重建归档
		}
		return
	}
	// 新建归档
	cache.Archives = append(cache.Archives, &model.Archive{
		Time:     article.CreatedAt,
		Articles: model.SortedArticles{article},
	})
	if needSort { // 重建归档
		PagesCh <- PageArchive
	}
}

// redelArticle 从tag、series、archive删除文章
func (cache *Cache) redelArticle(article *model.Article) {
	// tag
	for _, tag := range article.Tags {
		for i, v := range cache.TagArticles[tag] {
			if v == article {
				cache.TagArticles[tag] = append(cache.TagArticles[tag][0:i], cache.TagArticles[tag][i+1:]...)
				if len(cache.TagArticles[tag]) == 0 {
					delete(cache.TagArticles, tag)
				}
			}
		}
	}
	// serie
	for i, serie := range cache.Series {
		if serie.ID == article.SerieID {
			for j, v := range serie.Articles {
				if v == article {
					cache.Series[i].Articles = append(cache.Series[i].Articles[0:j],
						cache.Series[i].Articles[j+1:]...)
					PagesCh <- PageSeries
					break
				}
			}
		}
	}
	// archive
	for i, archive := range cache.Archives {
		ay, am, _ := archive.Time.Date()
		if y, m, _ := article.CreatedAt.Date(); ay == y && am == m {
			for j, v := range archive.Articles {
				if v == article {
					cache.Archives[i].Articles = append(cache.Archives[i].Articles[0:j],
						cache.Archives[i].Articles[j+1:]...)
					if len(cache.Archives[i].Articles) == 0 {
						cache.Archives = append(cache.Archives[:i], cache.Archives[i+1:]...)
					}
					PagesCh <- PageArchive
					break
				}
			}
		}
	}
}

// loadOrInit 读取数据或初始化
func (cache *Cache) loadOrInit() error {
	// blogger
	blogger := &model.Blogger{
		BlogName:  strings.Title(config.Conf.Account.Username),
		SubTitle:  "Rome was not built in one day.",
		BeiAn:     "蜀ICP备xxxxxxxx号-1",
		BTitle:    fmt.Sprintf("%s's Blog", strings.Title(config.Conf.Account.Username)),
		Copyright: `本站使用「<a href="//creativecommons.org/licenses/by/4.0/">署名 4.0 国际</a>」创作共享协议，转载请注明作者及原网址。`,
	}
	created, err := Store.LoadInsertBlogger(context.Background(), blogger)
	if err != nil {
		return err
	}
	cache.Blogger = blogger
	if created { // init articles: about blogroll
		about := &model.Article{
			ID:        1, // 固定ID
			Author:    config.Conf.Account.Username,
			Title:     "关于",
			Slug:      "about",
			CreatedAt: time.Time{}.AddDate(0, 0, 1),
		}
		err = Store.InsertArticle(context.Background(), about, ArticleStartID)
		if err != nil {
			return err
		}
		// 推送到 disqus
		go DisqusClient.ThreadCreate(about, blogger.BTitle)
		blogroll := &model.Article{
			ID:        2, // 固定ID
			Author:    config.Conf.Account.Username,
			Title:     "友情链接",
			Slug:      "blogroll",
			CreatedAt: time.Time{}.AddDate(0, 0, 7),
		}
		err = Store.InsertArticle(context.Background(), blogroll, ArticleStartID)
		if err != nil {
			return err
		}
	}
	// account
	pwd := tools.EncryptPasswd(config.Conf.Account.Username, config.Conf.Account.Password)

	account := &model.Account{
		Username: config.Conf.Account.Username,
		Password: pwd,
	}
	_, err = Store.LoadInsertAccount(context.Background(), account)
	if err != nil {
		return err
	}
	cache.Account = account
	// series
	series, err := Store.LoadAllSerie(context.Background())
	if err != nil {
		return err
	}
	cache.Series = series
	// all articles
	search := store.SearchArticles{
		Page:   1,
		Limit:  9999,
		Fields: map[string]interface{}{store.SearchArticleDraft: false},
	}
	articles, _, err := Store.LoadArticleList(context.Background(), search)
	if err != nil {
		return err
	}
	for i, v := range articles {
		// 渲染页面
		GenerateExcerptMarkdown(v)

		cache.ArticlesMap[v.Slug] = v
		// 分析文章
		if v.ID < ArticleStartID {
			continue
		}
		if i > 0 {
			v.Prev = articles[i-1]
		}
		if i < len(articles)-1 &&
			articles[i+1].ID >= ArticleStartID {
			v.Next = articles[i+1]
		}
		cache.readdArticle(v, false)
	}
	cache.Articles = articles
	// 重建专题与归档
	PagesCh <- PageSeries
	PagesCh <- PageArchive
	return nil
}

// regeneratePages 重新生成series,archive页面
func (cache *Cache) regeneratePages() {
	for {
		switch page := <-PagesCh; page {
		case PageSeries:
			sort.Sort(cache.Series)
			buf := bytes.Buffer{}
			buf.WriteString(cache.Blogger.SeriesSay)
			buf.WriteString("\n\n")
			for _, series := range cache.Series {
				buf.WriteString(fmt.Sprintf("### %s{#toc-%d}", series.Name, series.ID))
				buf.WriteByte('\n')
				buf.WriteString(series.Desc)
				buf.WriteString("\n\n")
				for _, article := range series.Articles {
					//eg. * [标题一](/post/hello-world.html) <span class="date">(Man 02, 2006)</span>
					str := fmt.Sprintf("* [%s](/post/%s.html) <span class=\"date\">(%s)</span>\n",
						article.Title, article.Slug, article.CreatedAt.Format("Jan 02, 2006"))
					buf.WriteString(str)
				}
				buf.WriteString("\n")
			}
			cache.PageSeries = string(PageRender(buf.Bytes()))
		case PageArchive:
			sort.Sort(cache.Archives)
			buf := bytes.Buffer{}
			buf.WriteString(cache.Blogger.ArchivesSay + "\n")
			var (
				currentYear string
				gt12Month   = len(cache.Archives) > 12
			)
			for _, archive := range cache.Archives {
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
						str := fmt.Sprintf("* *[%s](/post/%s.html) <span class=\"date\">(%s)</span>*\n",
							article.Title, article.Slug, createdAt.Format("Jan 02, 2006"))
						buf.WriteString(str)
					} else {
						str := fmt.Sprintf("* [%s](/post/%s.html) <span class=\"date\">(%s)</span>\n",
							article.Title, article.Slug, createdAt.Format("Jan 02, 2006"))
						buf.WriteString(str)
					}
				}
			}
			cache.PageArchives = string(PageRender(buf.Bytes()))
		}
	}
}
