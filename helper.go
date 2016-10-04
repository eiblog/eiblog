package main

import (
	"crypto/sha256"
	"fmt"
	"io"
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

func IgnoreHtmlTag(src string) string {
	//去除所有尖括号内的HTML代码
	re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
	src = re.ReplaceAllString(src, "")

	//去除换行符
	re, _ = regexp.Compile("\\s{1,}")
	return re.ReplaceAllString(src, "")
}
