// Package main provides ...
package main

import "time"

type Account struct {
	// 账户名
	Username string
	// 账户密码
	Password string
	// 二次验证token
	Token string
	// 账户
	Email string
	// 手机号
	PhoneN string
	// 住址
	Address string
	// 创建时间
	CreateTime time.Time
	// 最后登录时间
	LoginTime time.Time
	// 登出时间
	LogoutTime time.Time
	// 最后登录ip
	LoginIP string
	// 博客信息
	Blogger
}

type Blogger struct {
	// 博客名
	BlogName string
	// SubTitle
	SubTitle string
	// 备案号
	BeiAn string
	// 底部title
	BTitle string
	// 版权声明
	Copyright string
	// 专题，倒序
	SeriesSay string
	Series    SortSeries
	// 归档描述
	ArchivesSay string
	Archives    SortArchives
	// 忽略存储，前端界面全部缓存
	PageSeries   string                  `bson:"-"` // 专题页面
	PageArchives string                  `bson:"-"` // 归档页面
	Tags         map[string]SortArticles `bson:"-"` // 标签 name->tag
	Articles     SortArticles            `bson:"-"` // 所有文章
	MapArticles  map[string]*Article     `bson:"-"` // url->Article
	CH           chan string             `bson:"-"` // channel
}

type Serie struct {
	// 自增id
	ID int32
	// 名称unique
	Name string
	// 缩略名
	Slug string
	// 专题描述
	Desc string
	// 创建时间
	CreateTime time.Time
	// 文章
	Articles SortArticles `bson:"-"`
}

type SortSeries []*Serie

func (s SortSeries) Len() int           { return len(s) }
func (s SortSeries) Less(i, j int) bool { return s[i].ID > s[j].ID }
func (s SortSeries) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type Archive struct {
	Time     time.Time
	Articles SortArticles `bson:"-"`
}

type SortArchives []*Archive

func (s SortArchives) Len() int           { return len(s) }
func (s SortArchives) Less(i, j int) bool { return s[i].Time.After(s[j].Time) }
func (s SortArchives) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type Article struct {
	// 自增id
	ID int32
	// 作者名
	Author string
	// 标题
	Title string
	// 文章名: how-to-get-girlfriend
	Slug string
	// 评论数量
	Count int
	// markdown文档
	Content string
	// 归属专题
	SerieID int32
	// tagname
	Tags []string
	// 是否是草稿
	IsDraft bool
	// 创建时间
	CreateTime time.Time
	// 更新时间
	UpdateTime time.Time
	// 开始删除时间
	DeleteTime time.Time
	// 上篇文章
	Prev *Article `bson:"-"`
	// 下篇文章
	Next *Article `bson:"-"`
	// Header
	Header string `bson:"-"`
	// 预览信息
	Excerpt string `bson:"-"`
	// 一句话描述，文章第一句
	Desc string `bson:"-"`
	// disqus thread
	Thread string `bson:"-"`
}

type SortArticles []*Article

func (s SortArticles) Len() int           { return len(s) }
func (s SortArticles) Less(i, j int) bool { return s[i].CreateTime.After(s[j].CreateTime) }
func (s SortArticles) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
