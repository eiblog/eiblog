// Package model provides ...
package model

import "time"

// Article 文章
type Article struct {
	ID      int32  `gorm:"primaryKey;autoIncrement"` // 自增ID
	Author  string `gorm:"not null"`                 // 作者名
	Slug    string `gorm:"not null;uniqueIndex"`     // 文章缩略名
	Title   string `gorm:"not null"`                 // 标题
	Count   int    `gorm:"not null"`                 // 评论数量
	Content string `gorm:"not null"`                 // markdown内容
	SerieID int32  `gorm:"not null"`                 // 专题ID
	Tags    string `gorm:"not null"`                 // tag,以逗号隔开
	IsDraft bool   `gorm:"not null"`                 // 是否是草稿

	DeleteTime time.Time `gorm:"default:null"`  // 删除时间
	UpdateTime time.Time `gorm:"default:now()"` // 更新时间
	CreateTime time.Time `gorm:"default:now()"` // 创建时间

	Header  string   `gorm:"-" bson:"-"` // header
	Excerpt string   `gorm:"-" bson:"-"` // 预览信息
	Desc    string   `gorm:"-" bson:"-"` // 描述
	Thread  string   `gorm:"-" bson:"-"` // disqus thread
	Prev    *Article `gorm:"-" bson:"-"` // 上篇文章
	Next    *Article `gorm:"-" bson:"-"` // 下篇文章
}

// SortedArticles 按时间排序后文章
type SortedArticles []*Article

// Len 长度
func (s SortedArticles) Len() int { return len(s) }

// Less 对比
func (s SortedArticles) Less(i, j int) bool { return s[i].CreateTime.After(s[j].CreateTime) }

// Swap 交换
func (s SortedArticles) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
