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

// UploadParams upload params
type UploadParams struct {
	Name string
	Size int64
	Data io.Reader

	Conf config.Qiniu
}

// QiniuUpload 上传文件
func QiniuUpload(params UploadParams) (string, error) {
	if params.Conf.AccessKey == "" ||
		params.Conf.SecretKey == "" {
		return "", errors.New("qiniu config error")
	}
	key := completeQiniuKey(params.Name)

	mac := qbox.NewMac(params.Conf.AccessKey,
		params.Conf.SecretKey)
	// 设置上传策略
	putPolicy := &storage.PutPolicy{
		Scope:      params.Conf.Bucket,
		Expires:    3600,
		InsertOnly: 1,
	}
	// 上传token
	uploadToken := putPolicy.UploadToken(mac)
	// 上传配置
	cfg := &storage.Config{
		UseHTTPS: true,
	}
	// uploader
	uploader := storage.NewFormUploader(cfg)
	ret := new(storage.PutRet)
	putExtra := &storage.PutExtra{}

	err := uploader.Put(context.Background(), ret, uploadToken,
		key, params.Data, params.Size, putExtra)
	if err != nil {
		return "", err
	}
	url := "https://" + params.Conf.Domain + "/" + key
	return url, nil
}

// DeleteParams delete params
type DeleteParams struct {
	Name string
	Days int

	Conf config.Qiniu
}

// QiniuDelete 删除文件
func QiniuDelete(params DeleteParams) error {
	key := completeQiniuKey(params.Name)

	mac := qbox.NewMac(params.Conf.AccessKey,
		params.Conf.SecretKey)
	// 上传配置
	cfg := &storage.Config{
		Zone:     &storage.ZoneHuadong,
		UseHTTPS: true,
	}
	// manager
	bucketManager := storage.NewBucketManager(mac, cfg)
	// Delete
	if params.Days > 0 {
		return bucketManager.DeleteAfterDays(params.Conf.Bucket, key, params.Days)
	}
	return bucketManager.Delete(params.Conf.Bucket, key)
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
