package main

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/eiblog/eiblog/setting"
	"qiniupkg.com/api.v7/kodo"
	"qiniupkg.com/api.v7/kodocli"
	url "qiniupkg.com/x/url.v7"
)

var qiniu_cfg = &kodo.Config{
	AccessKey: setting.Conf.Kodo.AccessKey,
	SecretKey: setting.Conf.Kodo.SecretKey,
	Scheme:    "https",
}

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

// 上传文件
func FileUpload(name string, size int64, data io.Reader) (string, error) {
	if setting.Conf.Kodo.AccessKey == "" || setting.Conf.Kodo.SecretKey == "" {
		return "", errors.New("qiniu config error")
	}

	// 创建一个client
	c := kodo.New(0, qiniu_cfg)

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

	key := getKey(name)
	if key == "" {
		return "", errors.New("不支持的文件类型")
	}

	var ret PutRet
	var extra = kodocli.PutExtra{OnProgress: onProgress}
	err := uploader.Put(nil, &ret, token, key, data, size, &extra)
	if err != nil {
		return "", err
	}

	url := "https://" + setting.Conf.Kodo.Domain + "/" + url.Escape(key)
	return url, nil
}

// 删除文件
func FileDelete(name string) error {
	// new一个Bucket管理对象
	c := kodo.New(0, qiniu_cfg)
	p := c.Bucket(setting.Conf.Kodo.Name)

	key := getKey(name)
	if key == "" {
		return errors.New("不支持的文件类型")
	}

	// 调用Delete方法删除文件
	err := p.Delete(nil, key)
	// 打印返回值以及出错信息
	if err != nil {
		return err
	}
	return nil
}

func getKey(name string) string {
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
	}
	return key
}
