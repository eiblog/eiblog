package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
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

	localFile := "/Users/jemy/Documents/github.png"
	key := "qiniu-x.png"

	putPolicy := storage.PutPolicy{
		Scope: bucket,
	}
	mac := qbox.NewMac(accessKey, secretKey)

	cfg := storage.Config{}
	// 空间对应的机房
	cfg.Zone = &storage.ZoneHuadong
	// 是否使用https域名
	cfg.UseHTTPS = false
	// 上传是否使用CDN上传加速
	cfg.UseCdnDomains = false

	//设置代理
	proxyURL := "http://localhost:8888"
	proxyURI, _ := url.Parse(proxyURL)

	//绑定网卡
	nicIP := "100.100.33.138"
	dialer := &net.Dialer{
		LocalAddr: &net.TCPAddr{
			IP: net.ParseIP(nicIP),
		},
	}

	//构建代理client对象
	client := http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURI),
			Dial:  dialer.Dial,
		},
	}

	resumeUploader := storage.NewResumeUploaderEx(&cfg, &rpc.Client{Client: &client})
	upToken := putPolicy.UploadToken(mac)

	ret := storage.PutRet{}

	err := resumeUploader.PutFile(context.Background(), &ret, upToken, key, localFile, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(ret.Key, ret.Hash)
}
