package main

import (
	"regexp"
	"strconv"
	"time"
)

func CheckEmail(e string) bool {
	reg := regexp.MustCompile(`^(\w)+([\.\-]\w+)*#(\w)+((\.\w+)+)$`)
	return reg.MatchString(e)
}

func CheckDomain(domain string) bool {
	reg := regexp.MustCompile(`^(http://|https://)?[0-9a-zA-Z]+[0-9a-zA-Z\.-]*\.[a-zA-Z]{2,4}$`)
	return reg.MatchString(domain)
}

func CheckSMS(sms string) bool {
	reg := regexp.MustCompile(`^\+\d+$`)
	return reg.MatchString(sms)
}

func CheckPwd(pwd string) bool {
	return len(pwd) > 5 && len(pwd) < 19
}

func CheckDate(date string) time.Time {
	if t, err := time.Parse("2006-01-02 15:04", date); err == nil {
		return t
	}
	return time.Now()
}

func CheckSerieID(sid string) int32 {
	if id, err := strconv.Atoi(sid); err == nil {
		return int32(id)
	}
	return 0
}

func CheckBool(str string) bool {
	return str == "true"
}

func CheckPublish(do string) bool {
	return do == "publish"
}
