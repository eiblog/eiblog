// Package model provides ...
package model

import "time"

// Series 专题
type Series struct {
	ID         int32     `gorm:"primaryKey;autoIncrement"` // 自增ID
	Slug       string    `gorm:"not null;uniqueIndex"`     // 缩略名
	Name       string    `gorm:"not null"`                 // 专题名
	Desc       string    `gorm:"not null"`                 // 专题描述
	CreateTime time.Time `gorm:"default:now()"`            // 创建时间

	Articles SortedArticles `gorm:"-" bson:"-"` // 专题下的文章
}

// SortedSeries 排序后专题
type SortedSeries []*Series

// Len 长度
func (s SortedSeries) Len() int { return len(s) }

// Less 比较
func (s SortedSeries) Less(i, j int) bool { return s[i].ID > s[j].ID }

// Swap 交换
func (s SortedSeries) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
