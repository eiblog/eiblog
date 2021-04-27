// Package model provides ...
package model

import "time"

// use snake_case as column name

// Archive 归档
type Archive struct {
	Time time.Time `gorm:"column:time;not null" bson:"time"`

	Articles SortedArticles `gorm:"-" bson:"-"` // 归档下的文章
}

// SortedArchives 排序后的归档
type SortedArchives []*Archive

// Len 长度
func (s SortedArchives) Len() int { return len(s) }

// Less 比较
func (s SortedArchives) Less(i, j int) bool { return s[i].Time.After(s[j].Time) }

// Swap 交换
func (s SortedArchives) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
