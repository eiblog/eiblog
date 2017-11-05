package main

import (
	"fmt"
	"os"

	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/cdn"
)

var (
	accessKey = os.Getenv("QINIU_ACCESS_KEY")
	secretKey = os.Getenv("QINIU_SECRET_KEY")
	domain    = os.Getenv("QINIU_TEST_DOMAIN")
)

func main() {
	mac := qbox.NewMac(accessKey, secretKey)
	cdnManager := cdn.NewCdnManager(mac)

	// 预取链接，单次请求链接不可以超过100个，如果超过，请分批发送请求
	urlsToPrefetch := []string{
		"http://if-pbl.qiniudn.com/qiniu.png",
		"http://if-pbl.qiniudn.com/github.png",
	}
	ret, err := cdnManager.PrefetchUrls(urlsToPrefetch)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(ret.Code)
	fmt.Println(ret.RequestID)
}
