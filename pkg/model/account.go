// Package model provides ...
package model

import "time"

// Account 博客账户
type Account struct {
	Username string `gorm:"primaryKey"` // 用户名
	Password string `gorm:"not null"`   // 密码
	Email    string `gorm:"not null"`   // 邮件地址
	PhoneN   string `gorm:"not null"`   // 手机号
	Address  string `gorm:"not null"`   // 地址信息

	LogoutTime time.Time `gorm:"default:null"`  // 登出时间
	LoginIP    string    `gorm:"default:null"`  // 最近登录IP
	LoginUA    string    `gorm:"default:null"`  // 最近登录IP
	LoginTime  time.Time `gorm:"default:now()"` // 最近登录时间
	CreateTime time.Time `gorm:"default:now()"` // 创建时间
}
