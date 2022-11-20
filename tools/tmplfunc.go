// Package tools provides ...
package tools

import (
	"encoding/base64"
	htmpl "html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"text/template"
	"time"
)

var (
	// TplFuncMap template func map
	TplFuncMap = make(template.FuncMap)
	// TimeLocation set location timezone
	TimeLocation = time.UTC
)

func init() {
	TplFuncMap["dateformat"] = DateFormat
	TplFuncMap["str2html"] = Str2html
	TplFuncMap["join"] = Join
	TplFuncMap["isnotzero"] = IsNotZero
	TplFuncMap["getavatar"] = GetAvatar
}

// Str2html string to html
func Str2html(raw string) htmpl.HTML {
	return htmpl.HTML(raw)
}

// DateFormat takes a time and a layout string and returns a string with the formatted date.
// Used by the template parser as "dateformat"
func DateFormat(t time.Time, layout string) string {
	return t.In(TimeLocation).Format(layout)
}

// Join join string array with sep
func Join(a []string, sep string) string {
	return strings.Join(a, sep)
}

// IsNotZero judge t is zero
func IsNotZero(t time.Time) bool {
	return !t.IsZero()
}

// cache avatar image
// url: https://<static_domain>/static/img/avatar.png
var avatar string

// GetAvatar store avatar base64 into css
func GetAvatar(domain string) string {
	if avatar == "" {
		resp, err := http.Get("https://" + domain + "/static/img/avatar.png")
		if err != nil {
			log.Println(err)
			return ""
		}
		defer resp.Body.Close()

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			return ""
		}

		avatar = "data:" + resp.Header.Get("content-type") + ";base64," + base64.StdEncoding.EncodeToString(data)
	}

	return avatar
}

// ImgToNormal replace lazy image attr data-src to src
func ImgToNormal(content string) string {
	return strings.ReplaceAll(content, "data-src=", "src=")
}
