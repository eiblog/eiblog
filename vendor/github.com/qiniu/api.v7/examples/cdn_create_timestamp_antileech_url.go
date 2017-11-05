package main

import (
	"fmt"
	"time"

	"github.com/qiniu/api.v7/cdn"
)

func main() {
	urlStr := "http://image.example.com/qiniu_do_not_delete.gif"
	cryptKey := "your crypt key"
	deadline := time.Now().Add(time.Second * 3600).Unix()
	accessUrl, err := cdn.CreateTimestampAntileechURL(urlStr, cryptKey, deadline)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(accessUrl)
}
