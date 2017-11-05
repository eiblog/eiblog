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
	mac := qbox.NewMac(accessKey, secretKey)

	cfg := storage.Config{
		// 是否使用https域名进行资源管理
		UseHTTPS: false,
	}
	// 指定空间所在的区域，如果不指定将自动探测
	// 如果没有特殊需求，默认不需要指定
	//cfg.Zone=&storage.ZoneHuabei
	bucketManager := storage.NewBucketManager(mac, &cfg)

	srcBucket := "if-pbl"
	srcKey := "github.png"
	//目标空间可以和源空间相同，但是不能为跨机房的空间
	destBucket := srcBucket
	//目标文件名可以和源文件名相同，也可以不同
	destKey := "github-new.png"
	//如果目标文件存在，是否强制覆盖，如果不覆盖，默认返回614 file exists
	force := false
	err := bucketManager.Copy(srcBucket, srcKey, destBucket, destKey, force)
	if err != nil {
		fmt.Println(err)
		return
	}
}
