// Package model provides ...
package model

import "time"

// use snake_case as column name

// Account 博客账户
type Account struct {
	Username string `gorm:"column:username;primaryKey" bson:"username"` // 用户名
	Password string `gorm:"column:password;not null" bson:"password"`   // 密码
	Email    string `gorm:"column:email;not null" bson:"email"`         // 邮件地址
	PhoneN   string `gorm:"column:phone_n;not null" bson:"phone_n"`     // 手机号
	Address  string `gorm:"column:address;not null" bson:"address"`     // 地址信息

	LogoutAt  time.Time `gorm:"column:logout_at;not null" bson:"logout_at"`                    // 登出时间
	LoginIP   string    `gorm:"column:login_ip;not null" bson:"login_ip"`                      // 最近登录IP
	LoginUA   string    `gorm:"column:login_ua;not null" bson:"login_ua"`                      // 最近登录IP
	LoginAt   time.Time `gorm:"column:login_at;default:current_timestamp" bson:"login_at"`     // 最近登录时间
	CreatedAt time.Time `gorm:"column:created_at;default:current_timestamp" bson:"created_at"` // 创建时间
}
