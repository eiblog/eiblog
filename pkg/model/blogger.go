// Package model provides ...
package model

// Blogger 博客信息
type Blogger struct {
	BlogName  string `gorm:"column:blog_name;not null" bson:"blog_name"` // 博客名
	SubTitle  string `gorm:"column:sub_title;not null" bson:"sub_title"` // 子标题
	BeiAn     string `gorm:"column:bei_an;not null" bson:"bei_an"`       // 备案号
	BTitle    string `gorm:"column:b_title;not null" bson:"b_title"`     // 底部title
	Copyright string `gorm:"column:copyright;not null" bson:"copyright"` // 版权声明

	SeriesSay   string `gorm:"column:series_say;not null" bson:"series_say"`     // 专题说明
	ArchivesSay string `gorm:"column:archives_say;not null" bson:"archives_say"` // 归档说明
}
