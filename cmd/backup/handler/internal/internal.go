package internal

import (
	"github.com/eiblog/eiblog/cmd/backup/config"
	"github.com/eiblog/eiblog/pkg/third/qiniu"
)

// QiniuClient 七牛客户端
var QiniuClient *qiniu.QiniuClient

func init() {
	var err error
	QiniuClient, err = qiniu.NewQiniuClient(config.Conf.Qiniu)
	if err != nil {
		panic(err)
	}
}
