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
)

func main() {
	mac := qbox.NewMac(accessKey, secretKey)
	cfg := storage.Config{
		UseHTTPS: false,
	}
	// 指定空间所在的区域，如果不指定将自动探测
	// 如果没有特殊需求，默认不需要指定
	//cfg.Zone=&storage.ZoneHuabei
	operationManager := storage.NewOperationManager(mac, &cfg)
	persistentId := "z0.597f28b445a2650c994bb208"
	ret, err := operationManager.Prefop(persistentId)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(ret.String())
}
