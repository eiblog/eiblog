package main

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"path/filepath"

	"github.com/eiblog/eiblog/setting"
	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
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

// 进度条
func onProgress(fsize, uploaded int64) {
	d := int(float64(uploaded) / float64(fsize) * 100)
	if fsize == uploaded {
		fmt.Printf("\rUpload completed!          \n")
	} else {
		fmt.Printf("\r%02d%% uploaded              ", int(d))
	}
}

// 上传文件
func FileUpload(name string, size int64, data io.Reader) (string, error) {
	if setting.Conf.Qiniu.AccessKey == "" || setting.Conf.Qiniu.SecretKey == "" {
		return "", errors.New("qiniu config error")
	}

	key := getKey(name)
	if key == "" {
		return "", errors.New("不支持的文件类型")
	}

	mac := qbox.NewMac(setting.Conf.Qiniu.AccessKey, setting.Conf.Qiniu.SecretKey)
	// 设置上传的策略
	putPolicy := &storage.PutPolicy{
		Scope:      setting.Conf.Qiniu.Bucket,
		Expires:    3600,
		InsertOnly: 1,
	}
	// 上传token
	upToken := putPolicy.UploadToken(mac)

	// 上传配置
	cfg := &storage.Config{
		Zone:     &storage.ZoneHuadong,
		UseHTTPS: true,
	}
	// uploader
	uploader := storage.NewFormUploader(cfg)
	ret := new(storage.PutRet)
	putExtra := &storage.PutExtra{OnProgress: onProgress}
	err := uploader.Put(nil, ret, upToken, key, data, size, putExtra)
	if err != nil {
		return "", err
	}

	url := "https://" + setting.Conf.Qiniu.Domain + "/" + url.QueryEscape(key)
	return url, nil
}

// 删除文件
func FileDelete(name string) error {
	key := getKey(name)
	if key == "" {
		return errors.New("不支持的文件类型")
	}

	mac := qbox.NewMac(setting.Conf.Qiniu.AccessKey, setting.Conf.Qiniu.SecretKey)
	// 上传配置
	cfg := &storage.Config{
		Zone:     &storage.ZoneHuadong,
		UseHTTPS: true,
	}
	// manager
	bucketManager := storage.NewBucketManager(mac, cfg)
	// Delete
	err := bucketManager.Delete(setting.Conf.Qiniu.Bucket, key)
	if err != nil {
		return err
	}
	return nil
}

// 修复路径
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
