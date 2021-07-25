// Package model provides ...
package model

import "time"

// Serie 专题
type Serie struct {
	ID        int       `gorm:"column:id;primaryKey" bson:"id"`                                // 自增ID
	Slug      string    `gorm:"column:slug;not null;uniqueIndex" bson:"slug"`                  // 缩略名
	Name      string    `gorm:"column:name;not null" bson:"name"`                              // 专题名
	Desc      string    `gorm:"column:desc;not null" bson:"desc"`                              // 专题描述
	CreatedAt time.Time `gorm:"column:created_at;default:current_timestamp" bson:"created_at"` // 创建时间

	Articles SortedArticles `gorm:"-" bson:"-"` // 专题下的文章
}

// SortedSeries 排序后专题
type SortedSeries []*Serie

// Len 长度
func (s SortedSeries) Len() int { return len(s) }

// Less 比较
func (s SortedSeries) Less(i, j int) bool { return s[i].ID > s[j].ID }

// Swap 交换
func (s SortedSeries) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
