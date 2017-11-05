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

	//刷新链接，单次请求链接不可以超过100个，如果超过，请分批发送请求
	urlsToRefresh := []string{
		"http://if-pbl.qiniudn.com/qiniu.png",
		"http://if-pbl.qiniudn.com/github.png",
	}
	ret, err := cdnManager.RefreshUrls(urlsToRefresh)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(ret.Code)
	fmt.Println(ret.RequestID)

	// 刷新目录，刷新目录需要联系七牛技术支持开通权限
	// 单次请求链接不可以超过10个，如果超过，请分批发送请求
	dirsToRefresh := []string{
		"http://if-pbl.qiniudn.com/images/",
		"http://if-pbl.qiniudn.com/static/",
	}
	ret, err = cdnManager.RefreshDirs(dirsToRefresh)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(ret.Code)
	fmt.Println(ret.RequestID)
	fmt.Println(ret.Error)
}
