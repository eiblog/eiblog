// Package internal provides ...
package internal

import (
	"context"
	"errors"
	"io"
	"path/filepath"

	"github.com/eiblog/eiblog/pkg/config"

	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
)

// QiniuUpload 上传文件
func QiniuUpload(name string, size int64, data io.Reader) (string, error) {
	if config.Conf.EiBlogApp.Qiniu.AccessKey == "" ||
		config.Conf.EiBlogApp.Qiniu.SecretKey == "" {
		return "", errors.New("qiniu config error")
	}
	key := completeQiniuKey(name)

	mac := qbox.NewMac(config.Conf.EiBlogApp.Qiniu.AccessKey,
		config.Conf.EiBlogApp.Qiniu.SecretKey)
	// 设置上传策略
	putPolicy := &storage.PutPolicy{
		Scope:      config.Conf.EiBlogApp.Qiniu.Bucket,
		Expires:    3600,
		InsertOnly: 1,
	}
	// 上传token
	uploadToken := putPolicy.UploadToken(mac)
	// 上传配置
	cfg := &storage.Config{
		Zone:     &storage.ZoneHuadong,
		UseHTTPS: true,
	}
	// uploader
	uploader := storage.NewFormUploader(cfg)
	ret := new(storage.PutRet)
	putExtra := &storage.PutExtra{}

	err := uploader.Put(context.Background(), ret, uploadToken,
		key, data, size, putExtra)
	if err != nil {
		return "", err
	}
	url := "https://" + config.Conf.EiBlogApp.Qiniu.Domain + "/" + key
	return url, nil
}

// QiniuDelete 删除文件
func QiniuDelete(name string) error {
	key := completeQiniuKey(name)

	mac := qbox.NewMac(config.Conf.EiBlogApp.Qiniu.AccessKey,
		config.Conf.EiBlogApp.Qiniu.SecretKey)
	// 上传配置
	cfg := &storage.Config{
		Zone:     &storage.ZoneHuadong,
		UseHTTPS: true,
	}
	// manager
	bucketManager := storage.NewBucketManager(mac, cfg)
	// Delete
	return bucketManager.Delete(config.Conf.EiBlogApp.Qiniu.Bucket, key)
}

// completeQiniuKey 修复路径
func completeQiniuKey(name string) string {
	ext := filepath.Ext(name)

	switch ext {
	case ".bmp", ".png", ".jpg",
		".gif", ".ico", ".jpeg":

		name = "blog/img/" + name
	case ".mov", ".mp4":
		name = "blog/video/" + name
	case ".go", ".js", ".css",
		".cpp", ".php", ".rb",
		".java", ".py", ".sql",
		".lua", ".html", ".sh",
		".xml", ".cs":

		name = "blog/code/" + name
	case ".txt", ".md", ".ini",
		".yaml", ".yml", ".doc",
		".ppt", ".pdf":

		name = "blog/document/" + name
	case ".zip", ".rar", ".tar",
		".gz":

		name = "blog/archive/" + name
	default:
		name = "blog/other/" + name
	}
	return name
}
