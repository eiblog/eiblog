// Package model provides ...
package model

// Blogger 博客信息
type Blogger struct {
	BlogName  string `gorm:"not null"` // 博客名
	SubTitle  string `gorm:"not null"` // 子标题
	BeiAn     string `gorm:"not null"` // 备案号
	BTitle    string `gorm:"not null"` // 底部title
	Copyright string `gorm:"not null"` // 版权声明

	SeriesSay   string `gorm:"not null"` // 专题说明
	ArchivesSay string `gorm:"not null"` // 归档说明
}
