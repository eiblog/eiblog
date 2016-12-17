package main

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/eiblog/eiblog/setting"
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

func Upload(name string, size int64, data io.Reader) (string, error) {
	if setting.Conf.Kodo.AccessKey == "" || setting.Conf.Kodo.SecretKey == "" {
		return "", errors.New("qiniu config error")
	}

	conf.ACCESS_KEY = setting.Conf.Kodo.AccessKey
	conf.SECRET_KEY = setting.Conf.Kodo.SecretKey
	// 创建一个client
	c := kodo.New(0, nil)

	// 设置上传的策略
	policy := &kodo.PutPolicy{
		Scope:      setting.Conf.Kodo.Name,
		Expires:    3600,
		InsertOnly: 1,
	}

	// 生成一个上传token
	token := c.MakeUptoken(policy)
	// 构建一个uploader
	zone := 0
	uploader := kodocli.NewUploader(zone, nil)

	ext := filepath.Ext(name)
	var key string
	switch ext {
	case ".bmp", ".png", ".jpg", ".gif", ".ico":
		key = "blog/img/" + name
	case ".mov", ".mp4":
		key = "blog/video/" + name
	case ".go", ".js", ".css", ".cpp", ".php", ".rb", ".java", ".py", ".sql", ".lua", ".html", ".sh", ".xml", ".cs":
		key = "blog/code/" + name
	case ".txt", ".md", ".ini", ".yaml", ".yml", ".doc", ".ppt", ".pdf":
		key = "blog/document/" + name
	case ".zip", ".rar", ".tar", ".gz":
		key = "blog/archive/" + name
	default:
		return "", errors.New("不支持的文件类型")
	}

	var ret PutRet
	var extra = kodocli.PutExtra{OnProgress: onProgress}
	err := uploader.Put(nil, &ret, token, key, data, size, &extra)
	if err != nil {
		return "", err
	}

	url := kodo.MakeBaseUrl(setting.Conf.Kodo.Domain, ret.Key)
	return url, nil
}
