// Package main provides ...
package main

import (
	"os"
	"testing"
)

func TestUpload(t *testing.T) {
	path := "/Users/chen/Desktop/png-MicroService-by-StuQ.png"
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	info, _ := file.Stat()
	url, err := FileUpload(info.Name(), info.Size(), file)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(url)
}
