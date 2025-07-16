package qiniu

import (
	"context"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"time"

	"github.com/eiblog/eiblog/pkg/config"

	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
)

// QiniuClient qiniu client
type QiniuClient struct {
	Conf config.Qiniu
}

// NewQiniuClient new qiniu client
func NewQiniuClient(conf config.Qiniu) (*QiniuClient, error) {
	if conf.AccessKey == "" ||
		conf.SecretKey == "" ||
		conf.Bucket == "" ||
		conf.Domain == "" {
		return nil, errors.New("qiniu config error")
	}
	return &QiniuClient{Conf: conf}, nil
}

// UploadParams upload params
type UploadParams struct {
	Name           string
	Size           int64
	Data           io.Reader
	NoCompletePath bool
}

// Upload 上传文件
func (cli *QiniuClient) Upload(params UploadParams) (string, error) {
	key := params.Name
	if !params.NoCompletePath {
		key = filepath.Base(params.Name)
	}

	mac := qbox.NewMac(cli.Conf.AccessKey,
		cli.Conf.SecretKey)
	// 设置上传策略
	putPolicy := &storage.PutPolicy{
		Scope:      cli.Conf.Bucket,
		Expires:    3600,
		InsertOnly: 1,
	}
	// 上传token
	uploadToken := putPolicy.UploadToken(mac)
	// 上传配置
	region, err := storage.GetRegion(cli.Conf.AccessKey, cli.Conf.Bucket)
	if err != nil {
		return "", err
	}
	cfg := &storage.Config{
		UseHTTPS: true,
		Region:   region,
	}
	// uploader
	uploader := storage.NewFormUploader(cfg)
	ret := new(storage.PutRet)
	putExtra := &storage.PutExtra{}

	err = uploader.Put(context.Background(), ret, uploadToken,
		key, params.Data, params.Size, putExtra)
	if err != nil {
		return "", err
	}
	url := "https://" + cli.Conf.Domain + "/" + key
	return url, nil
}

// DeleteParams delete params
type DeleteParams struct {
	Name           string
	Days           int
	NoCompletePath bool
}

// QiniuDelete 删除文件
func (cli *QiniuClient) Delete(params DeleteParams) error {
	key := params.Name
	if !params.NoCompletePath {
		key = completeQiniuKey(params.Name)
	}

	mac := qbox.NewMac(cli.Conf.AccessKey,
		cli.Conf.SecretKey)
	// 上传配置
	region, err := storage.GetRegion(cli.Conf.AccessKey, cli.Conf.Bucket)
	if err != nil {
		return err
	}
	cfg := &storage.Config{
		UseHTTPS: true,
		Region:   region,
	}
	// manager
	bucketManager := storage.NewBucketManager(mac, cfg)
	// Delete
	if params.Days > 0 {
		return bucketManager.DeleteAfterDays(cli.Conf.Bucket, key, params.Days)
	}
	return bucketManager.Delete(cli.Conf.Bucket, key)
}

// ContentParams list params
type ContentParams struct {
	Prefix string
}

// Content 获取文件内容
func (cli *QiniuClient) Content(params ContentParams) ([]byte, error) {
	mac := qbox.NewMac(cli.Conf.AccessKey,
		cli.Conf.SecretKey)
	// region
	region, err := storage.GetRegion(cli.Conf.AccessKey, cli.Conf.Bucket)
	if err != nil {
		return nil, err
	}
	cfg := &storage.Config{
		UseHTTPS: true,
		Region:   region,
	}
	// manager
	bucketManager := storage.NewBucketManager(mac, cfg)
	// list file
	files, _, _, _, err := bucketManager.ListFiles(cli.Conf.Bucket, params.Prefix, "", "", 1)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, errors.New("no file")
	}
	deadline := time.Now().Add(time.Second * 60).Unix()
	url := storage.MakePrivateURLv2(mac, "https://"+cli.Conf.Domain, files[0].Key, deadline)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// completeQiniuKey 修复路径
func completeQiniuKey(name string) string {
	ext := filepath.Ext(name)

	switch ext {
	case ".bmp", ".png", ".jpg", ".gif", ".ico", ".jpeg":
		name = "blog/img/" + name
	case ".mov", ".mp4":
		name = "blog/video/" + name
	case ".go", ".js", ".css", ".cpp", ".php", ".rb", ".java",
		".py", ".sql", ".lua", ".html", ".sh", ".xml", ".cs":
		name = "blog/code/" + name
	case ".txt", ".md", ".ini", ".yaml", ".yml", ".doc", ".ppt", ".pdf":
		name = "blog/document/" + name
	case ".zip", ".rar", ".tar", ".gz":
		name = "blog/archive/" + name
	default:
		name = "blog/other/" + name
	}
	return name
}
