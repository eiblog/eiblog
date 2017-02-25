// Package main provides ...
package main

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/eiblog/blackfriday"
	"github.com/eiblog/eiblog/setting"
	"github.com/eiblog/utils/logd"
	db "github.com/eiblog/utils/mgo"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// 数据库及表名
const (
	DB                 = "eiblog"
	COLLECTION_ACCOUNT = "account"
	COLLECTION_ARTICLE = "article"
	COUNTER_SERIE      = "serie"
	COUNTER_ARTICLE    = "article"
	SERIES_MD          = "series_md"
	ARCHIVE_MD         = "archive_md"
	ADD                = "add"
	DELETE             = "delete"
)

// blackfriday 配置
const (
	commonHtmlFlags = 0 |
		blackfriday.HTML_TOC |
		blackfriday.HTML_USE_XHTML |
		blackfriday.HTML_USE_SMARTYPANTS |
		blackfriday.HTML_SMARTYPANTS_FRACTIONS |
		blackfriday.HTML_SMARTYPANTS_DASHES |
		blackfriday.HTML_SMARTYPANTS_LATEX_DASHES |
		blackfriday.HTML_NOFOLLOW_LINKS

	commonExtensions = 0 |
		blackfriday.EXTENSION_NO_INTRA_EMPHASIS |
		blackfriday.EXTENSION_TABLES |
		blackfriday.EXTENSION_FENCED_CODE |
		blackfriday.EXTENSION_AUTOLINK |
		blackfriday.EXTENSION_STRIKETHROUGH |
		blackfriday.EXTENSION_SPACE_HEADERS |
		blackfriday.EXTENSION_HEADER_IDS |
		blackfriday.EXTENSION_BACKSLASH_LINE_BREAK |
		blackfriday.EXTENSION_DEFINITION_LISTS
)

// Global Account
var (
	Ei   *Account
	lock sync.Mutex
)

func init() {
	// 数据库加索引
	ms, c := db.Connect(DB, COLLECTION_ACCOUNT)
	index := mgo.Index{
		Key:        []string{"username"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	if err := c.EnsureIndex(index); err != nil {
		logd.Fatal(err)
	}
	ms.Close()
	ms, c = db.Connect(DB, COLLECTION_ARTICLE)
	index = mgo.Index{
		Key:        []string{"id"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	if err := c.EnsureIndex(index); err != nil {
		logd.Fatal(err)
	}
	index = mgo.Index{
		Key:        []string{"slug"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	if err := c.EnsureIndex(index); err != nil {
		logd.Fatal(err)
	}
	ms.Close()
	// 读取帐号信息
	Ei = loadAccount()
	// 获取文章数据
	Ei.Articles = loadArticles()
	// 生成markdown文档
	go generateMarkdown()
	// 启动定时器
	go timer()
	// 获取评论数量
	go PostsCount()
}

// 读取或初始化帐号信息
func loadAccount() (a *Account) {
	a = &Account{}
	err := db.FindOne(DB, COLLECTION_ACCOUNT, bson.M{"username": setting.Conf.Account.Username}, a)
	// 初始化用户数据
	if err == mgo.ErrNotFound {
		a = &Account{
			Username:   setting.Conf.Account.Username,
			Password:   EncryptPasswd(setting.Conf.Account.Username, setting.Conf.Account.Password),
			Email:      setting.Conf.Account.Email,
			PhoneN:     setting.Conf.Account.PhoneNumber,
			Address:    setting.Conf.Account.Address,
			CreateTime: time.Now(),
		}
		a.BlogName = setting.Conf.Blogger.BlogName
		a.SubTitle = setting.Conf.Blogger.SubTitle
		a.BeiAn = setting.Conf.Blogger.BeiAn
		a.BTitle = setting.Conf.Blogger.BTitle
		a.Copyright = setting.Conf.Blogger.Copyright
		err = db.Insert(DB, COLLECTION_ACCOUNT, a)
		generateTopic()
	} else if err != nil {
		logd.Fatal(err)
	}
	a.CH = make(chan string, 2)
	a.MapArticles = make(map[string]*Article)
	a.Tags = make(map[string]SortArticles)
	return
}

func loadArticles() (artcs SortArticles) {
	err := db.FindAll(DB, COLLECTION_ARTICLE, bson.M{"isdraft": false, "deletetime": bson.M{"$eq": time.Time{}}}, &artcs)
	if err != nil {
		logd.Fatal(err)
	}
	sort.Sort(artcs)
	for i, v := range artcs {
		// 渲染文章
		GenerateExcerptAndRender(v)
		Ei.MapArticles[v.Slug] = v
		// 分析文章
		if v.ID < setting.Conf.StartID {
			continue
		}
		if i > 0 {
			v.Prev = artcs[i-1]
		}
		if artcs[i+1].ID >= setting.Conf.StartID {
			v.Next = artcs[i+1]
		}
		ManageTagsArticle(v, false, ADD)
		ManageSeriesArticle(v, false, ADD)
		ManageArchivesArticle(v, false, ADD)
	}
	Ei.CH <- SERIES_MD
	Ei.CH <- ARCHIVE_MD
	return
}

// generate series,archive markdown
func generateMarkdown() {
	for {
		switch typ := <-Ei.CH; typ {
		case SERIES_MD:
			sort.Sort(Ei.Series)
			var buffer bytes.Buffer
			buffer.WriteString(Ei.SeriesSay)
			buffer.WriteString("\n\n")
			for _, serie := range Ei.Series {
				buffer.WriteString(fmt.Sprintf("### %s{#toc-%d}", serie.Name, serie.ID))
				buffer.WriteString("\n")
				buffer.WriteString(serie.Desc)
				buffer.WriteString("\n\n")
				for _, artc := range serie.Articles {
					//eg. * [标题一](/post/hello-world.html) <span class="date">(Man 02, 2006)</span>
					buffer.WriteString("* [" + artc.Title + "](/post/" + artc.Slug + ".html) <span class=\"date\">(" + artc.CreateTime.Format("Jan 02, 2006") + ")</span>\n")
				}
				buffer.WriteByte('\n')
			}
			Ei.PageSeries = string(renderPage(buffer.Bytes()))
		case ARCHIVE_MD:
			sort.Sort(Ei.Archives)
			var buffer bytes.Buffer
			buffer.WriteString(Ei.ArchivesSay)
			buffer.WriteString("\n\n")
			for _, archive := range Ei.Archives {
				buffer.WriteString(fmt.Sprintf("### %s", archive.Time.Format("2006年01月")))
				buffer.WriteString("\n\n")
				for _, artc := range archive.Articles {
					buffer.WriteString("* [" + artc.Title + "](/post/" + artc.Slug + ".html) <span class=\"date\">(" + artc.CreateTime.Format("Jan 02, 2006") + ")</span>\n")
				}
				buffer.WriteByte('\n')
			}
			Ei.PageArchives = string(renderPage(buffer.Bytes()))
		}
	}
}

// init account: generate blogroll and about page
func generateTopic() {
	about := &Article{
		ID:         db.NextVal(DB, COUNTER_ARTICLE),
		Author:     setting.Conf.Account.Username,
		Title:      "关于",
		Slug:       "about",
		CreateTime: time.Now(),
		UpdateTime: time.Time{},
	}
	blogroll := &Article{
		ID:         db.NextVal(DB, COUNTER_ARTICLE),
		Author:     setting.Conf.Account.Username,
		Title:      "友情链接",
		Slug:       "blogroll",
		UpdateTime: time.Now(),
		CreateTime: time.Time{},
	}
	err := db.Insert(DB, COLLECTION_ARTICLE, blogroll)
	if err != nil {
		logd.Fatal(err)
	}
	err = db.Insert(DB, COLLECTION_ARTICLE, about)
	if err != nil {
		logd.Fatal(err)
	}
}

// render page
func renderPage(md []byte) []byte {
	renderer := blackfriday.HtmlRenderer(commonHtmlFlags, "", "")
	return blackfriday.Markdown(md, renderer, commonExtensions)
}

// 文章分页
func PageList(p, n int) (prev int, next int, artcs []*Article) {
	var l int
	for l = len(Ei.Articles); l > 0; l-- {
		if Ei.Articles[l-1].ID >= setting.Conf.StartID {
			break
		}
	}
	if l == 0 {
		return 0, 0, nil
	}
	m := l / n
	if d := l % n; d > 0 {
		m++
	}
	if p > m {
		p = m
	}
	if p > 1 {
		prev = p - 1
	}
	if p < m {
		next = p + 1
	}
	s := (p - 1) * n
	e := p * n
	if e > l {
		e = l
	}
	artcs = Ei.Articles[s:e]
	return
}

func ManageTagsArticle(artc *Article, s bool, do string) {
	switch do {
	case ADD:
		for _, tag := range artc.Tags {
			Ei.Tags[tag] = append(Ei.Tags[tag], artc)
			if s {
				sort.Sort(Ei.Tags[tag])
			}
		}
	case DELETE:
		for _, tag := range artc.Tags {
			for i, v := range Ei.Tags[tag] {
				if v == artc {
					Ei.Tags[tag] = append(Ei.Tags[tag][0:i], Ei.Tags[tag][i+1:]...)
					if len(Ei.Tags[tag]) == 0 {
						delete(Ei.Tags, tag)
					}
					return
				}
			}
		}
	}
}

func ManageSeriesArticle(artc *Article, s bool, do string) {
	switch do {
	case ADD:
		for i, serie := range Ei.Series {
			if serie.ID == artc.SerieID {
				Ei.Series[i].Articles = append(Ei.Series[i].Articles, artc)
				if s {
					sort.Sort(Ei.Series[i].Articles)
					Ei.CH <- SERIES_MD
					return
				}
			}
		}
	case DELETE:
		for i, serie := range Ei.Series {
			if serie.ID == artc.SerieID {
				for j, v := range serie.Articles {
					if v == artc {
						Ei.Series[i].Articles = append(Ei.Series[i].Articles[0:j], Ei.Series[i].Articles[j+1:]...)
						Ei.CH <- SERIES_MD
						return
					}
				}
			}
		}
	}
}

func ManageArchivesArticle(artc *Article, s bool, do string) {
	switch do {
	case ADD:
		add := false
		y, m, _ := artc.CreateTime.Date()
		for i, archive := range Ei.Archives {
			ay, am, _ := archive.Time.Date()
			if y == ay && m == am {
				add = true
				Ei.Archives[i].Articles = append(Ei.Archives[i].Articles, artc)
				if s {
					sort.Sort(Ei.Archives[i].Articles)
					Ei.CH <- ARCHIVE_MD
					break
				}
			}
		}
		if !add {
			Ei.Archives = append(Ei.Archives, &Archive{Time: artc.CreateTime, Articles: SortArticles{artc}})
		}
	case DELETE:
		for i, archive := range Ei.Archives {
			ay, am, _ := archive.Time.Date()
			if y, m, _ := artc.CreateTime.Date(); ay == y && am == m {
				for j, v := range archive.Articles {
					if v == artc {
						Ei.Archives[i].Articles = append(Ei.Archives[i].Articles[0:j], Ei.Archives[i].Articles[j+1:]...)
						Ei.CH <- ARCHIVE_MD
						return
					}
				}
			}
		}
	}
}

// 渲染markdown操作和截取摘要操作
var reg = regexp.MustCompile(setting.Conf.Identifier)

// header
var regH = regexp.MustCompile("</nav></div>")

func GenerateExcerptAndRender(artc *Article) {
	if strings.HasPrefix(artc.Content, setting.Conf.Description) {
		index := strings.Index(artc.Content, "\r\n")
		artc.Desc = IgnoreHtmlTag(artc.Content[len(setting.Conf.Description):index])
		artc.Content = artc.Content[index:]
	}

	content := renderPage([]byte(artc.Content))
	index := regH.FindIndex(content)
	if index != nil {
		artc.Header = string(content[0:index[1]])
		artc.Content = string(content[index[1]:])
	} else {
		artc.Content = string(content)
	}
	index = reg.FindStringIndex(artc.Content)
	if index != nil {
		artc.Excerpt = IgnoreHtmlTag(artc.Content[0:index[0]])
	} else {
		uc := []rune(artc.Content)
		length := setting.Conf.Length
		if len(uc) < length {
			length = len(uc)
		}
		artc.Excerpt = IgnoreHtmlTag(string(uc[0:length]))
	}
}

// 读取草稿箱
func LoadDraft() (artcs SortArticles, err error) {
	err = db.FindAll(DB, COLLECTION_ARTICLE, bson.M{"isdraft": true}, &artcs)
	sort.Sort(artcs)
	return
}

// 读取回收箱
func LoadTrash() (artcs SortArticles, err error) {
	err = db.FindAll(DB, COLLECTION_ARTICLE, bson.M{"deletetime": bson.M{"$ne": time.Time{}}}, &artcs)
	sort.Sort(artcs)
	return
}

// 添加文章
func AddArticle(artc *Article) error {
	// 分配ID, 占位至起始id
	for {
		if id := db.NextVal(DB, COUNTER_ARTICLE); id < setting.Conf.StartID {
			continue
		} else {
			artc.ID = id
			break
		}
	}
	if !artc.IsDraft {
		defer GenerateExcerptAndRender(artc)
		Ei.MapArticles[artc.Slug] = artc
		Ei.Articles = append([]*Article{artc}, Ei.Articles...)
		sort.Sort(Ei.Articles)
		AddToLinkedList(artc.ID)
		ManageTagsArticle(artc, true, ADD)
		ManageSeriesArticle(artc, true, ADD)
		ManageArchivesArticle(artc, true, ADD)
		Ei.CH <- ARCHIVE_MD
		if artc.SerieID > 0 {
			Ei.CH <- SERIES_MD
		}
	}
	return db.Insert(DB, COLLECTION_ARTICLE, artc)
}

// 删除文章，移入回收箱
func DelArticles(ids ...int32) error {
	lock.Lock()
	defer lock.Unlock()
	for _, id := range ids {
		i, artc := GetArticle(id)
		DelFromLinkedList(artc)
		Ei.Articles = append(Ei.Articles[:i], Ei.Articles[i+1:]...)
		delete(Ei.MapArticles, artc.Slug)
		ManageTagsArticle(artc, false, DELETE)
		ManageSeriesArticle(artc, false, DELETE)
		ManageArchivesArticle(artc, false, DELETE)
		err := UpdateArticle(bson.M{"id": id}, bson.M{"$set": bson.M{"deletetime": time.Now()}})
		if err != nil {
			return err
		}
		artc = nil
	}
	Ei.CH <- ARCHIVE_MD
	Ei.CH <- SERIES_MD
	return nil
}

func DelFromLinkedList(artc *Article) {
	if artc.Prev == nil && artc.Next != nil {
		artc.Next.Prev = nil
	} else if artc.Prev != nil && artc.Next == nil {
		artc.Prev.Next = nil
	} else if artc.Prev != nil && artc.Next != nil {
		artc.Prev.Next = artc.Next
		artc.Next.Prev = artc.Prev
	}
}

func AddToLinkedList(id int32) {
	i, artc := GetArticle(id)
	if i == 0 && Ei.Articles[i+1].ID >= setting.Conf.StartID {
		artc.Next = Ei.Articles[i+1]
		Ei.Articles[i+1].Prev = artc
	} else if i > 0 && Ei.Articles[i-1].ID >= setting.Conf.StartID {
		artc.Prev = Ei.Articles[i-1]
		if Ei.Articles[i-1].Next != nil {
			artc.Next = Ei.Articles[i-1].Next
			Ei.Articles[i-1].Next.Prev = artc
		}
		Ei.Articles[i-1].Next = artc
	}
}

// 从缓存获取文章
func GetArticle(id int32) (int, *Article) {
	for i, artc := range Ei.Articles {
		if id == artc.ID {
			return i, artc
		}
	}
	return -1, nil
}

// 定时清除回收箱文章
func timer() {
	delT := time.NewTicker(time.Duration(setting.Conf.Clean) * time.Hour)
	for {
		<-delT.C
		db.Remove(DB, COLLECTION_ARTICLE, bson.M{"deletetime": bson.M{"$gt": time.Time{}, "$lt": time.Now().Add(time.Duration(setting.Conf.Trash) * time.Hour)}})
	}
}

// 操作帐号字段
func UpdateAccountField(M bson.M) error {
	return db.Update(DB, COLLECTION_ACCOUNT, bson.M{"username": Ei.Username}, M)
}

// 删除草稿箱或回收箱，永久删除
func RemoveArticle(id int32) error {
	return db.Remove(DB, COLLECTION_ARTICLE, bson.M{"id": id})
}

// 恢复删除文章到草稿箱
func RecoverArticle(id int32) error {
	return db.Update(DB, COLLECTION_ARTICLE, bson.M{"id": id}, bson.M{"$set": bson.M{"deletetime": time.Time{}, "isdraft": true}})
}

// 更新文章
func UpdateArticle(query, update interface{}) error {
	return db.Update(DB, COLLECTION_ARTICLE, query, update)
}

// 编辑文档
func QueryArticle(id int32) *Article {
	artc := &Article{}
	if err := db.FindOne(DB, COLLECTION_ARTICLE, bson.M{"id": id}, artc); err != nil {
		return nil
	}
	return artc
}

// 添加专题
func AddSerie(name, slug, desc string) error {
	serie := &Serie{db.NextVal(DB, COUNTER_SERIE), name, slug, desc, time.Now(), nil}
	Ei.Series = append(Ei.Series, serie)
	sort.Sort(Ei.Series)
	Ei.CH <- SERIES_MD
	return UpdateAccountField(bson.M{"$addToSet": bson.M{"blogger.series": serie}})
}

// 更新专题
func UpdateSerie(serie *Serie) error {
	Ei.CH <- SERIES_MD
	return db.Update(DB, COLLECTION_ACCOUNT, bson.M{"username": Ei.Username, "blogger.series.id": serie.ID}, bson.M{"$set": bson.M{"blogger.series.$": serie}})
}

// 删除专题
func DelSerie(id int32) error {
	for i, serie := range Ei.Series {
		if id == serie.ID {
			if len(serie.Articles) > 0 {
				return fmt.Errorf("请删除该专题下的所有文章")
			}
			err := UpdateAccountField(bson.M{"$pull": bson.M{"blogger.series": bson.M{"id": id}}})
			if err != nil {
				return err
			}
			Ei.Series[i] = nil
			Ei.Series = append(Ei.Series[:i], Ei.Series[i+1:]...)
			Ei.CH <- SERIES_MD
		}
	}
	return nil
}

// 查找专题
func QuerySerie(id int32) *Serie {
	for _, serie := range Ei.Series {
		if serie.ID == id {
			return serie
		}
	}
	return nil
}

// 后台分页
func PageListBack(se int, kw string, draft, del bool, p, n int) (max int, artcs []*Article) {
	M := bson.M{}
	if draft {
		M["isdraft"] = true
	} else if del {
		M["deletetime"] = bson.M{"$ne": time.Time{}}
	} else {
		M["isdraft"] = false
		M["deletetime"] = bson.M{"$eq": time.Time{}}
		if se > 0 {
			M["serieid"] = se
		}
		if kw != "" {
			M["title"] = bson.M{"$regex": kw, "$options": "$i"}
		}
	}
	ms, c := db.Connect(DB, COLLECTION_ARTICLE)
	defer ms.Close()
	err := c.Find(M).Select(bson.M{"content": 0}).Sort("-createtime").Limit(n).Skip((p - 1) * n).All(&artcs)
	if err != nil {
		logd.Error(err)
	}
	count, err := c.Find(M).Count()
	if err != nil {
		logd.Error(err)
	}
	max = count / n
	if count%n > 0 {
		max++
	}
	return
}
