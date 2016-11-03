package main

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/eiblog/utils/logd"
	"qiniupkg.com/api.v7/conf"
	"qiniupkg.com/api.v7/kodo"
	"qiniupkg.com/api.v7/kodocli"
)

type bucket struct {
	name      string
	domain    string
	accessKey string
	secretKey string
}

var buckets = map[string]*bucket{}

func getBucket(typ string) *bucket {
	return buckets[typ]
}

type PutRet struct {
	Hash string `json:"hash"`
	Key  string `json:"key"`
}

func onProgress(fsize, uploaded int64) {
	d := int(float64(uploaded) / float64(fsize) * 100)
	if fsize == uploaded {
		fmt.Printf("\rUpload completed!          ")
	} else {
		fmt.Printf("\r%02d%% uploaded              ", int(d))
	}
}

func upload(typ string, filepath string) {
	bucket := getBucket(typ)
	if bucket == nil {
		logd.Debug("invalid type:", typ)
		return
	}

	file, err := os.Open(filepath)
	if err != nil {
		logd.Debugf("%s", err.Error())
		return
	}
	data, err := ioutil.ReadAll(file)
	file.Close()
	chksum := fmt.Sprintf("%x", md5.Sum(data))
	ext := path.Ext(filepath)

	conf.ACCESS_KEY = bucket.accessKey
	conf.SECRET_KEY = bucket.secretKey
	// 创建一个client
	c := kodo.New(0, nil)

	// 设置上传的策略
	policy := &kodo.PutPolicy{
		Scope:      bucket.name,
		Expires:    3600,
		InsertOnly: 1,
	}

	// 生成一个上传token
	token := c.MakeUptoken(policy)
	// 构建一个uploader
	zone := 0
	uploader := kodocli.NewUploader(zone, nil)

	var ret PutRet
	key := fmt.Sprintf("%s-%s%s", typ, chksum, ext)

	fmt.Printf("Uploading .....")
	var extra = kodocli.PutExtra{OnProgress: onProgress}
	res := uploader.PutFile(nil, &ret, token, key, filepath, &extra)
	// 打印返回的信息
	if res != nil {
		logd.Debugf("failed to upload patch file: %v", res)
		return
	}

	url := kodo.MakeBaseUrl(bucket.domain, key)
	fmt.Printf("url: %s\n", url)
}
