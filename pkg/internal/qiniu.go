// Package internal provides ...
package internal

import (
	"context"
	"errors"
	"io"
	"path/filepath"

	"github.com/eiblog/eiblog/v2/pkg/config"

	"github.com/qiniu/api.v7/v7/auth/qbox"
	"github.com/qiniu/api.v7/v7/storage"
)

// QiniuUpload 上传文件
func QiniuUpload(name string, size int64, data io.Reader) (string, error) {
	if config.Conf.BlogApp.Qiniu.AccessKey == "" ||
		config.Conf.BlogApp.Qiniu.SecretKey == "" {
		return "", errors.New("qiniu config error")
	}
	key := completeQiniuKey(name)

	mac := qbox.NewMac(config.Conf.BlogApp.Qiniu.AccessKey,
		config.Conf.BlogApp.Qiniu.SecretKey)
	// 设置上传策略
	putPolicy := &storage.PutPolicy{
		Scope:      config.Conf.BlogApp.Qiniu.Bucket,
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
	url := "https://" + config.Conf.BlogApp.Qiniu.Domain + "/" + key
	return url, nil
}

// QiniuDelete 删除文件
func QiniuDelete(name string) error {
	key := completeQiniuKey(name)

	mac := qbox.NewMac(config.Conf.BlogApp.Qiniu.AccessKey,
		config.Conf.BlogApp.Qiniu.SecretKey)
	// 上传配置
	cfg := &storage.Config{
		Zone:     &storage.ZoneHuadong,
		UseHTTPS: true,
	}
	// manager
	bucketManager := storage.NewBucketManager(mac, cfg)
	// Delete
	return bucketManager.Delete(config.Conf.BlogApp.Qiniu.Bucket, key)
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
