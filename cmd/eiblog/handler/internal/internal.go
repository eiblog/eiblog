package internal

import (
	"github.com/eiblog/eiblog/cmd/eiblog/config"
	"github.com/eiblog/eiblog/pkg/third/disqus"
	"github.com/eiblog/eiblog/pkg/third/es"
	"github.com/eiblog/eiblog/pkg/third/pinger"
	"github.com/eiblog/eiblog/pkg/third/qiniu"

	"github.com/sirupsen/logrus"
)

var (
	ESClient     *es.ESClient
	DisqusClient *disqus.DisqusClient
	QiniuClient  *qiniu.QiniuClient
	Pinger       *pinger.Pinger
)

func init() {
	var err error
	ESClient, err = es.NewESClient(config.Conf.ESHost)
	if err != nil {
		logrus.Fatal("init es client: ", err)
	}
	DisqusClient, err = disqus.NewDisqusClient(config.Conf.Host, config.Conf.Disqus)
	if err != nil {
		logrus.Fatal("init disqus client: ", err)
	}
	QiniuClient, err = qiniu.NewQiniuClient(config.Conf.Qiniu)
	if err != nil {
		logrus.Fatal("init qiniu client: ", err)
	}
	Pinger, err = pinger.NewPinger(config.Conf.Host, config.Conf.FeedRPC)
	if err != nil {
		logrus.Fatal("init pinger: ", err)
	}
}
