package main

import (
	"regexp"
	"strconv"
	"time"
)

// 检查 email
func CheckEmail(e string) bool {
	reg := regexp.MustCompile(`^(\w)+([\.\-]\w+)*@(\w)+((\.\w+)+)$`)
	return reg.MatchString(e)
}

// 检查 domain
func CheckDomain(domain string) bool {
	reg := regexp.MustCompile(`^(http://|https://)?[0-9a-zA-Z]+[0-9a-zA-Z\.-]*\.[a-zA-Z]{2,4}$`)
	return reg.MatchString(domain)
}

// 检查 sms
func CheckSMS(sms string) bool {
	reg := regexp.MustCompile(`^\+\d+$`)
	return reg.MatchString(sms)
}

// 检查 password
func CheckPwd(pwd string) bool {
	return len(pwd) > 5 && len(pwd) < 19
}

// 检查日期
func CheckDate(date string) time.Time {
	if t, err := time.ParseInLocation("2006-01-02 15:04", date, time.Local); err == nil {
		return t
	}
	return time.Now()
}

// 检查 id
func CheckSerieID(sid string) int32 {
	if id, err := strconv.Atoi(sid); err == nil {
		return int32(id)
	}
	return 0
}

// bool
func CheckBool(str string) bool {
	return str == "true" || str == "1"
}
