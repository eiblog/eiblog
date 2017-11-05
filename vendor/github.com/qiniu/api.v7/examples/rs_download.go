package main

import (
	"fmt"
	"os"
	"time"

	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
)

var (
	accessKey = os.Getenv("QINIU_ACCESS_KEY")
	secretKey = os.Getenv("QINIU_SECRET_KEY")
	bucket    = os.Getenv("QINIU_TEST_BUCKET")
)

func main() {
	mac := qbox.NewMac(accessKey, secretKey)

	// 公开空间访问
	domain := "https://image.example.com"
	key := "这是一个测试文件.jpg"
	publicAccessURL := storage.MakePublicURL(domain, key)
	fmt.Println(publicAccessURL)

	// 私有空间访问
	domain = "https://image.example.com"
	key = "这是一个测试文件.jpg"
	deadline := time.Now().Add(time.Second * 3600).Unix() //1小时有效期
	privateAccessURL := storage.MakePrivateURL(mac, domain, key, deadline)
	fmt.Println(privateAccessURL)
}
