package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"regexp"
	"time"

	"github.com/eiblog/utils/logd"
	"github.com/eiblog/utils/uuid"
)

const (
	SUCCESS = iota
	FAIL
)

// encrypt password
func EncryptPasswd(name, pass string) string {
	salt := "%$@w*)("
	h := sha256.New()
	io.WriteString(h, name)
	io.WriteString(h, salt)
	io.WriteString(h, pass)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func VerifyPasswd(origin, name, input string) bool {
	return origin == EncryptPasswd(name, input)
}

func RandUUIDv4() string {
	return uuid.NewV4().String()
}

func ReadDir(dir string, filter func(name string) bool) (files []string) {
	fis, err := ioutil.ReadDir(dir)
	if err != nil {
		return
	}
	for _, fi := range fis {
		if filter(fi.Name()) {
			continue
		}
		if fi.IsDir() {
			files = append(files, ReadDir(path.Join(dir, fi.Name()), filter)...)
			continue
		}
		files = append(files, path.Join(dir, fi.Name()))
	}
	return
}

func IgnoreHtmlTag(src string) string {
	//去除所有尖括号内的HTML代码
	re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
	src = re.ReplaceAllString(src, "")

	//去除换行符
	re, _ = regexp.Compile("\\s{2,}")
	return re.ReplaceAllString(src, "")
}

func PickFirstImage(html string) string {
	re, _ := regexp.Compile(`data-src="(.*?)"`)
	sli := re.FindAllStringSubmatch(html, 1)
	if len(sli) > 0 && len(sli[0]) > 1 {
		return sli[0][1]
	}
	return ""
}

// 2016-10-22T07:03:01
const (
	JUST_NOW    = "几秒前"
	MINUTES_AGO = "%d分钟前"
	HOURS_AGO   = "%d小时前"
	DAYS_AGO    = "%d天前"
	MONTH_AGO   = "%d月前"
	YEARS_AGO   = "%d年前"
)

func ConvertStr(str string) string {
	t, err := time.ParseInLocation("2006-01-02T15:04:05", str, time.UTC)
	if err != nil {
		logd.Error(err, str)
		return JUST_NOW
	}
	now := time.Now()
	year1, month1, day1 := t.Date()
	year2, month2, day2 := now.UTC().Date()
	if y := year2 - year1; y > 0 {
		return fmt.Sprintf(YEARS_AGO, y)
	}
	if m := month2 - month1; m > 0 {
		return fmt.Sprintf(MONTH_AGO, m)
	}
	if d := day2 - day1; d > 0 {
		return fmt.Sprintf(DAYS_AGO, d)
	}
	hour1, minute1, _ := t.Clock()
	hour2, minute2, _ := now.Clock()
	if h := hour2 - hour1; h > 0 {
		return fmt.Sprintf(HOURS_AGO, h)
	}
	if m := minute2 - minute1; m > 0 {
		return fmt.Sprintf(MINUTES_AGO, m)
	}
	return JUST_NOW
}
