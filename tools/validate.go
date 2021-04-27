// Package tools provides ...
package tools

import (
	"regexp"
)

var regexpEmail = regexp.MustCompile(`^(\w)+([\.\-]\w+)*@(\w)+((\.\w+)+)$`)

// ValidateEmail 校验邮箱
func ValidateEmail(e string) bool {
	return regexpEmail.MatchString(e)
}

var regexpPhoneNo = regexp.MustCompile(`^\+\d+$`)

// ValidatePhoneNo 校验手机号
func ValidatePhoneNo(no string) bool {
	return regexpPhoneNo.MatchString(no)
}

// ValidatePassword 校验米阿莫
func ValidatePassword(pwd string) bool {
	return len(pwd) > 5 && len(pwd) < 32
}
