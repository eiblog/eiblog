package main

import (
	"fmt"
	"os"

	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
)

var (
	accessKey = os.Getenv("QINIU_ACCESS_KEY")
	secretKey = os.Getenv("QINIU_SECRET_KEY")
	bucket    = os.Getenv("QINIU_TEST_BUCKET")
)

func main() {
	cfg := storage.Config{}
	mac := qbox.NewMac(accessKey, secretKey)
	bucketManger := storage.NewBucketManager(mac, &cfg)
	siteURL := "http://devtools.qiniu.com"

	// 设置镜像存储
	err := bucketManger.SetImage(siteURL, bucket)
	if err != nil {
		fmt.Println(err)
	}

	// 取消设置镜像存储
	err = bucketManger.UnsetImage(bucket)
	if err != nil {
		fmt.Println(err)
	}

}
