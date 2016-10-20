package main

import (
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"regexp"
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

func SHA1(data []byte) [sha1.Size]byte {
	return sha1.Sum(data)
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
	re, _ = regexp.Compile("\\s{1,}")
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
