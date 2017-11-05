package main

import (
	"fmt"
	"os"

	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
	"github.com/qiniu/x/rpc.v7"
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
	chgmKeys := map[string]string{
		"github1.png": "image/x-png",
		"github2.png": "image/x-png",
		"github3.png": "image/x-png",
		"github4.png": "image/x-png",
		"github5.png": "image/x-png",
	}
	chgmOps := make([]string, 0, len(chgmKeys))
	for key, newMime := range chgmKeys {
		chgmOps = append(chgmOps, storage.URIChangeMime(bucket, key, newMime))
	}
	rets, err := bucketManager.Batch(chgmOps)
	if err != nil {
		// 遇到错误
		if _, ok := err.(*rpc.ErrorInfo); ok {
			for _, ret := range rets {
				// 200 为成功
				fmt.Printf("%d\n", ret.Code)
				if ret.Code != 200 {
					fmt.Printf("%s\n", ret.Data.Error)
				}
			}
		} else {
			fmt.Printf("batch error, %s", err)
		}
	} else {
		// 完全成功
		for _, ret := range rets {
			// 200 为成功
			fmt.Printf("%d\n", ret.Code)
			if ret.Code != 200 {
				fmt.Printf("%s\n", ret.Data.Error)
			}
		}
	}
}
