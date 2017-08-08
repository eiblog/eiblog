// Package main provides ...
package main

import (
	"os"
	"testing"
)

func TestUpload(t *testing.T) {
	path := "qiniu.go"
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	info, _ := file.Stat()
	url, err := FileUpload(info.Name(), info.Size(), file)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(url)
}
