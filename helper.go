package main

import (
	"crypto/sha256"
	"fmt"
	"io"
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
