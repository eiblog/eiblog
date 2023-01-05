// Package tools provides ...
package tools

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"path"
	"regexp"
	"time"
)

// EncryptPasswd encrypt password
func EncryptPasswd(name, pass string) string {
	salt := "%$@w*)("
	h := sha256.New()
	io.WriteString(h, name)
	io.WriteString(h, salt)
	io.WriteString(h, pass)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// ReadDirFiles 读取目录
func ReadDirFiles(dir string, filter func(fi fs.FileInfo) bool) (files []string) {
	fileInfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return
	}
	for _, fi := range fileInfos {
		if filter(fi) {
			continue
		}
		if fi.IsDir() {
			files = append(files, ReadDirFiles(path.Join(dir, fi.Name()), filter)...)
			continue
		}
		files = append(files, path.Join(dir, fi.Name()))
	}
	return
}

// 2016-10-22T07:03:01
const (
	JustNow    = "几秒前"
	MinutesAgo = "%d分钟前"
	HoursAgo   = "%d小时前"
	DaysAgo    = "%d天前"
	MonthAgo   = "%d月前"
	YearsAgo   = "%d年前"
)

// ConvertStr 时间转换为间隔
func ConvertStr(str string) string {
	t, err := time.ParseInLocation("2006-01-02T15:04:05", str, time.UTC)
	if err != nil {
		return JustNow
	}
	now := time.Now().UTC()
	y1, m1, d1 := t.Date()
	y2, m2, d2 := now.Date()
	h1, mi1, s1 := t.Clock()
	h2, mi2, s2 := now.Clock()
	if y := y2 - y1; y > 1 || (y == 1 && m2-m1 >= 0) {
		return fmt.Sprintf(YearsAgo, y)
	} else if m := y*12 + int(m2-m1); m > 1 || (m == 1 && d2-d1 >= 0) {
		return fmt.Sprintf(MonthAgo, m)
	} else if d := m*dayIn(y1, m1) + d2 - d1; d > 1 || (d == 1 && h2-h1 >= 0) {
		return fmt.Sprintf(DaysAgo, d)
	} else if h := d*24 + h2 - h1; h > 1 || (h == 1 && mi2-mi1 >= 0) {
		return fmt.Sprintf(HoursAgo, h)
	} else if mi := h*60 + mi2 - mi1; mi > 1 || (mi == 1 && s2-s1 >= 0) {
		return fmt.Sprintf(MinutesAgo, mi)
	}
	return JustNow
}

// dayIn 获取天数
func dayIn(year int, m time.Month) int {
	if m == time.February && isLeapYear(year) {
		return 29
	}
	return monthToDays[m]
}

// monthToDays 月份转换
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

// isLeapYear是否是闰年
func isLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

var regexpImg = regexp.MustCompile(`data-src="(.*?)"`)

// PickFirstImage 获取第一张图片
func PickFirstImage(html string) string {
	sli := regexpImg.FindAllStringSubmatch(html, 1)
	if len(sli) > 0 && len(sli[0]) > 1 {
		return sli[0][1]
	}
	return ""
}

var (
	regexpBrackets = regexp.MustCompile(`<[\S\s]+?>`)
	regexpEnter    = regexp.MustCompile(`\s+`)
)

// IgnoreHTMLTag 去掉 html tag
func IgnoreHTMLTag(src string) string {
	// 去除所有尖括号内的HTML代码
	src = regexpBrackets.ReplaceAllString(src, "")
	// 去除换行符
	return regexpEnter.ReplaceAllString(src, "")
}
