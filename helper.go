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

// 月份转换
var monthToDays = map[time.Month]int{
	time.January:   31,
	time.February:  28,
	time.March:     31,
	time.April:     30,
	time.May:       31,
	time.June:      30,
	time.July:      31,
	time.August:    31,
	time.September: 30,
	time.October:   31,
	time.November:  30,
	time.December:  31,
}

// encrypt password
func EncryptPasswd(name, pass string) string {
	salt := "%$@w*)("
	h := sha256.New()
	io.WriteString(h, name)
	io.WriteString(h, salt)
	io.WriteString(h, pass)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// 验证密码
func VerifyPasswd(origin, name, input string) bool {
	return origin == EncryptPasswd(name, input)
}

// 随机 uuid
func RandUUIDv4() string {
	return uuid.NewV4().String()
}

// 读取目录
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

// 去掉 html tag
func IgnoreHtmlTag(src string) string {
	// 去除所有尖括号内的HTML代码
	re, _ := regexp.Compile(`<[\S\s]+?>`)
	src = re.ReplaceAllString(src, "")

	// 去除换行符
	re, _ = regexp.Compile(`\s+`)
	return re.ReplaceAllString(src, "")
}

// 获取第一张图片
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

// 时间转换为间隔
func ConvertStr(str string) string {
	t, err := time.ParseInLocation("2006-01-02T15:04:05", str, time.UTC)
	if err != nil {
		logd.Error(err, str)
		return JUST_NOW
	}
	now := time.Now().UTC()
	y1, m1, d1 := t.Date()
	y2, m2, d2 := now.Date()
	h1, mi1, s1 := t.Clock()
	h2, mi2, s2 := now.Clock()
	if y := y2 - y1; y > 1 || (y == 1 && m2-m1 >= 0) {
		return fmt.Sprintf(YEARS_AGO, y)
	} else if m := y*12 + int(m2-m1); m > 1 || (m == 1 && d2-d1 >= 0) {
		return fmt.Sprintf(MONTH_AGO, m)
	} else if d := m*dayIn(y1, m1) + d2 - d1; d > 1 || (d == 1 && h2-h1 >= 0) {
		return fmt.Sprintf(DAYS_AGO, d)
	} else if h := d*24 + h2 - h1; h > 1 || (h == 1 && mi2-mi1 >= 0) {
		return fmt.Sprintf(HOURS_AGO, h)
	} else if mi := h*60 + mi2 - mi1; mi > 1 || (mi == 1 && s2-s1 >= 0) {
		return fmt.Sprintf(MINUTES_AGO, mi)
	}
	return JUST_NOW
}

// 获取天数
func dayIn(year int, m time.Month) int {
	if m == time.February && isLeap(year) {
		return 29
	}
	return monthToDays[m]
}

// 是否是闰年
func isLeap(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}
