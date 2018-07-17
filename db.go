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
	"github.com/eiblog/utils/mgo"
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
	err := mgo.Index(DB, COLLECTION_ACCOUNT, []string{"username"})
	if err != nil {
		logd.Fatal(err)
	}

	err = mgo.Index(DB, COLLECTION_ARTICLE, []string{"id"})
	if err != nil {
		logd.Fatal(err)
	}

	err = mgo.Index(DB, COLLECTION_ARTICLE, []string{"slug"})
	if err != nil {
		logd.Fatal(err)
	}
	// 读取帐号信息
	loadAccount()
	// 获取文章数据
	loadArticles()
	// 生成markdown文档
	go generateMarkdown()
	// 启动定时器
	go timer()
	// 获取评论数量
	go PostsCount()
}

// 读取或初始化帐号信息
func loadAccount() {
	Ei = &Account{}
	err := mgo.FindOne(DB, COLLECTION_ACCOUNT, mgo.M{"username": setting.Conf.Account.Username}, Ei)
	// 初始化用户数据
	if err == mgo.ErrNotFound {
		logd.Printf("Initializing account: %s\n", setting.Conf.Account.Username)
		Ei = &Account{
			Username:   setting.Conf.Account.Username,
			Password:   EncryptPasswd(setting.Conf.Account.Username, setting.Conf.Account.Password),
			Email:      setting.Conf.Account.Email,
			PhoneN:     setting.Conf.Account.PhoneNumber,
			Address:    setting.Conf.Account.Address,
			CreateTime: time.Now(),
		}
		Ei.BlogName = setting.Conf.Blogger.BlogName
		Ei.SubTitle = setting.Conf.Blogger.SubTitle
		Ei.BeiAn = setting.Conf.Blogger.BeiAn
		Ei.BTitle = setting.Conf.Blogger.BTitle
		Ei.Copyright = setting.Conf.Blogger.Copyright
		err = mgo.Insert(DB, COLLECTION_ACCOUNT, Ei)
		generateTopic()
	} else if err != nil {
		logd.Fatal(err)
	}
	Ei.CH = make(chan string, 2)
	Ei.MapArticles = make(map[string]*Article)
	Ei.Tags = make(map[string]SortArticles)
}

func loadArticles() {
	err := mgo.FindAll(DB, COLLECTION_ARTICLE, mgo.M{"isdraft": false, "deletetime": mgo.M{"$eq": time.Time{}}}, &Ei.Articles)
	if err != nil {
		logd.Fatal(err)
	}
	sort.Sort(Ei.Articles)
	for i, v := range Ei.Articles {
		// 渲染文章
		GenerateExcerptAndRender(v)
		Ei.MapArticles[v.Slug] = v
		// 分析文章
		if v.ID < setting.Conf.General.StartID {
			continue
		}
		if i > 0 {
			v.Prev = Ei.Articles[i-1]
		}
		if Ei.Articles[i+1].ID >= setting.Conf.General.StartID {
			v.Next = Ei.Articles[i+1]
		}
		upArticle(v, false)
	}
	Ei.CH <- SERIES_MD
	Ei.CH <- ARCHIVE_MD
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
					buffer.WriteString("* [" + artc.Title + "](/post/" + artc.Slug +
						".html) <span class=\"date\">(" + artc.CreateTime.Format("Jan 02, 2006") + ")</span>\n")
				}
				buffer.WriteByte('\n')
			}
			Ei.PageSeries = string(renderPage(buffer.Bytes()))
		case ARCHIVE_MD:
			sort.Sort(Ei.Archives)
			var buffer bytes.Buffer
			buffer.WriteString(Ei.ArchivesSay + "\n")

			var (
				currentYear string
				gt12Month   = len(Ei.Archives) > 12
			)
			for _, archive := range Ei.Archives {
				if gt12Month {
					year := archive.Time.Format("2006 年")
					if currentYear != year {
						currentYear = year
						buffer.WriteString(fmt.Sprintf("\n### %s\n\n", archive.Time.Format("2006 年")))
					}
				} else {
					buffer.WriteString(fmt.Sprintf("\n### %s\n\n", archive.Time.Format("2006年1月")))
				}
				for i, artc := range archive.Articles {
					if i == 0 && gt12Month {
						buffer.WriteString("* *[" + artc.Title + "](/post/" + artc.Slug +
							".html) <span class=\"date\">(" + artc.CreateTime.Format("Jan 02, 2006") + ")</span>*\n")
					} else {
						buffer.WriteString("* [" + artc.Title + "](/post/" + artc.Slug +
							".html) <span class=\"date\">(" + artc.CreateTime.Format("Jan 02, 2006") + ")</span>\n")
					}
				}
			}
			Ei.PageArchives = string(renderPage(buffer.Bytes()))
		}
	}
}

// init account: generate blogroll and about page
func generateTopic() {
	about := &Article{
		ID:         mgo.NextVal(DB, COUNTER_ARTICLE),
		Author:     setting.Conf.Account.Username,
		Title:      "关于",
		Slug:       "about",
		CreateTime: time.Time{},
		UpdateTime: time.Time{},
	}
	// 推送到 disqus
	go func() { ThreadCreate(about) }()

	blogroll := &Article{
		ID:         mgo.NextVal(DB, COUNTER_ARTICLE),
		Author:     setting.Conf.Account.Username,
		Title:      "友情链接",
		Slug:       "blogroll",
		CreateTime: time.Time{},
		UpdateTime: time.Time{},
	}
	err := mgo.Insert(DB, COLLECTION_ARTICLE, blogroll)
	if err != nil {
		logd.Fatal(err)
	}
	err = mgo.Insert(DB, COLLECTION_ARTICLE, about)
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
		if Ei.Articles[l-1].ID >= setting.Conf.General.StartID {
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

// 渲染markdown操作和截取摘要操作
var reg = regexp.MustCompile(setting.Conf.General.Identifier)

// header
var regH = regexp.MustCompile("</nav></div>")

func GenerateExcerptAndRender(artc *Article) {
	if strings.HasPrefix(artc.Content, setting.Conf.General.DescPrefix) {
		index := strings.Index(artc.Content, "\r\n")
		artc.Desc = IgnoreHtmlTag(artc.Content[len(setting.Conf.General.DescPrefix):index])
		artc.Content = artc.Content[index:]
	}

	// 查找目录
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
		length := setting.Conf.General.Length
		if len(uc) < length {
			length = len(uc)
		}
		artc.Excerpt = IgnoreHtmlTag(string(uc[0:length]))
	}
}

// 读取草稿箱
func LoadDraft() (artcs SortArticles, err error) {
	err = mgo.FindAll(DB, COLLECTION_ARTICLE, mgo.M{"isdraft": true}, &artcs)
	sort.Sort(artcs)
	return
}

// 读取回收箱
func LoadTrash() (artcs SortArticles, err error) {
	err = mgo.FindAll(DB, COLLECTION_ARTICLE, mgo.M{"deletetime": mgo.M{"$ne": time.Time{}}}, &artcs)
	sort.Sort(artcs)
	return
}

// 添加文章到tag、serie、archive
func upArticle(artc *Article, needSort bool) {
	// tag
	for _, tag := range artc.Tags {
		Ei.Tags[tag] = append(Ei.Tags[tag], artc)
		if needSort {
			sort.Sort(Ei.Tags[tag])
		}
	}
	// serie
	for i, serie := range Ei.Series {
		if serie.ID == artc.SerieID {
			Ei.Series[i].Articles = append(Ei.Series[i].Articles, artc)
			if needSort {
				sort.Sort(Ei.Series[i].Articles)
				Ei.CH <- SERIES_MD
			}
			break
		}
	}
	// archive
	y, m, _ := artc.CreateTime.Date()
	for i, archive := range Ei.Archives {
		if ay, am, _ := archive.Time.Date(); y == ay && m == am {
			Ei.Archives[i].Articles = append(Ei.Archives[i].Articles, artc)
			if needSort {
				sort.Sort(Ei.Archives[i].Articles)
				Ei.CH <- ARCHIVE_MD
			}
			return
		}
	}
	Ei.Archives = append(Ei.Archives, &Archive{Time: artc.CreateTime,
		Articles: SortArticles{artc}})
	if needSort {
		Ei.CH <- ARCHIVE_MD
	}
}

// 删除文章从tag、serie、archive
func dropArticle(artc *Article) {
	// tag
	for _, tag := range artc.Tags {
		for i, v := range Ei.Tags[tag] {
			if v == artc {
				Ei.Tags[tag] = append(Ei.Tags[tag][0:i], Ei.Tags[tag][i+1:]...)
				if len(Ei.Tags[tag]) == 0 {
					delete(Ei.Tags, tag)
				}
			}
		}
	}
	// serie
	for i, serie := range Ei.Series {
		if serie.ID == artc.SerieID {
			for j, v := range serie.Articles {
				if v == artc {
					Ei.Series[i].Articles = append(Ei.Series[i].Articles[0:j],
						Ei.Series[i].Articles[j+1:]...)
					Ei.CH <- SERIES_MD
					break
				}
			}
		}
	}
	// archive
	for i, archive := range Ei.Archives {
		ay, am, _ := archive.Time.Date()
		if y, m, _ := artc.CreateTime.Date(); ay == y && am == m {
			for j, v := range archive.Articles {
				if v == artc {
					Ei.Archives[i].Articles = append(Ei.Archives[i].Articles[0:j],
						Ei.Archives[i].Articles[j+1:]...)
					if len(Ei.Archives[i].Articles) == 0 {
						Ei.Archives = append(Ei.Archives[:i], Ei.Archives[i+1:]...)
					}
					Ei.CH <- ARCHIVE_MD
					break
				}
			}
		}
	}
}

// 替换文章
func ReplaceArticle(oldArtc *Article, newArtc *Article) {
	Ei.MapArticles[newArtc.Slug] = newArtc
	GenerateExcerptAndRender(newArtc)
	if newArtc.ID < setting.Conf.General.StartID {
		return
	}
	if oldArtc != nil {
		i, artc := GetArticle(oldArtc.ID)
		DelFromLinkedList(artc)
		Ei.Articles = append(Ei.Articles[:i], Ei.Articles[i+1:]...)

		dropArticle(oldArtc)
	}

	Ei.Articles = append(Ei.Articles, newArtc)
	sort.Sort(Ei.Articles)
	AddToLinkedList(newArtc.ID)

	upArticle(newArtc, true)
}

// 添加文章
func AddArticle(artc *Article) error {
	// 分配ID, 占位至起始id
	for {
		if id := mgo.NextVal(DB, COUNTER_ARTICLE); id < setting.Conf.General.StartID {
			continue
		} else {
			artc.ID = id
			break
		}
	}

	err := mgo.Insert(DB, COLLECTION_ARTICLE, artc)
	if err != nil {
		return err
	}

	// 正式发布文章
	if !artc.IsDraft {
		defer GenerateExcerptAndRender(artc)
		Ei.MapArticles[artc.Slug] = artc
		Ei.Articles = append([]*Article{artc}, Ei.Articles...)
		sort.Sort(Ei.Articles)
		AddToLinkedList(artc.ID)

		upArticle(artc, true)
	}
	return nil
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

		err := UpdateArticle(mgo.M{"id": id}, mgo.M{"$set": mgo.M{"deletetime": time.Now()}})
		if err != nil {
			return err
		}
		dropArticle(artc)
	}
	return nil
}

// 从链表里删除文章
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

// 将文章添加到链表
func AddToLinkedList(id int32) {
	i, artc := GetArticle(id)
	if i == 0 && Ei.Articles[i+1].ID >= setting.Conf.General.StartID {
		artc.Next = Ei.Articles[i+1]
		Ei.Articles[i+1].Prev = artc
	} else if i > 0 && Ei.Articles[i-1].ID >= setting.Conf.General.StartID {
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
	delT := time.NewTicker(time.Duration(setting.Conf.General.Clean) * time.Hour)
	for {
		<-delT.C
		mgo.Remove(DB, COLLECTION_ARTICLE, mgo.M{"deletetime": mgo.M{"$gt": time.Time{},
			"$lt": time.Now().Add(time.Duration(setting.Conf.General.Trash) * time.Hour)}})
	}
}

// 操作帐号字段
func UpdateAccountField(M mgo.M) error {
	return mgo.Update(DB, COLLECTION_ACCOUNT, mgo.M{"username": Ei.Username}, M)
}

// 删除草稿箱或回收箱，永久删除
func RemoveArticle(id int32) error {
	return mgo.Remove(DB, COLLECTION_ARTICLE, mgo.M{"id": id})
}

// 恢复删除文章到草稿箱
func RecoverArticle(id int32) error {
	return mgo.Update(DB, COLLECTION_ARTICLE, mgo.M{"id": id},
		mgo.M{"$set": mgo.M{"deletetime": time.Time{}, "isdraft": true}})
}

// 更新文章
func UpdateArticle(query, update interface{}) error {
	return mgo.Update(DB, COLLECTION_ARTICLE, query, update)
}

// 编辑文档
func QueryArticle(id int32) *Article {
	artc := &Article{}
	if err := mgo.FindOne(DB, COLLECTION_ARTICLE, mgo.M{"id": id}, artc); err != nil {
		return nil
	}
	return artc
}

// 添加专题
func AddSerie(name, slug, desc string) error {
	serie := &Serie{mgo.NextVal(DB, COUNTER_SERIE), name, slug, desc, time.Now(), nil}
	Ei.Series = append(Ei.Series, serie)
	sort.Sort(Ei.Series)
	Ei.CH <- SERIES_MD
	return UpdateAccountField(mgo.M{"$addToSet": mgo.M{"blogger.series": serie}})
}

// 更新专题
func UpdateSerie(serie *Serie) error {
	Ei.CH <- SERIES_MD
	return mgo.Update(DB, COLLECTION_ACCOUNT, mgo.M{"username": Ei.Username,
		"blogger.series.id": serie.ID}, mgo.M{"$set": mgo.M{"blogger.series.$": serie}})
}

// 删除专题
func DelSerie(id int32) error {
	for i, serie := range Ei.Series {
		if id == serie.ID {
			if len(serie.Articles) > 0 {
				return fmt.Errorf("请删除该专题下的所有文章")
			}
			err := UpdateAccountField(mgo.M{"$pull": mgo.M{"blogger.series": mgo.M{"id": id}}})
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
	M := mgo.M{}
	if draft {
		M["isdraft"] = true
	} else if del {
		M["deletetime"] = mgo.M{"$ne": time.Time{}}
	} else {
		M["isdraft"] = false
		M["deletetime"] = mgo.M{"$eq": time.Time{}}
		if se > 0 {
			M["serieid"] = se
		}
		if kw != "" {
			M["title"] = mgo.M{"$regex": kw, "$options": "$i"}
		}
	}
	ms, c := mgo.Connect(DB, COLLECTION_ARTICLE)
	defer ms.Close()
	err := c.Find(M).Select(mgo.M{"content": 0}).Sort("-createtime").Limit(n).Skip((p - 1) * n).All(&artcs)
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
