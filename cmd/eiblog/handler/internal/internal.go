package internal

import (
	"io/fs"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/eiblog/eiblog/cmd/eiblog/config"
	"github.com/eiblog/eiblog/cmd/eiblog/handler/internal/store"
	"github.com/eiblog/eiblog/pkg/third/disqus"
	"github.com/eiblog/eiblog/pkg/third/es"
	"github.com/eiblog/eiblog/pkg/third/pinger"
	"github.com/eiblog/eiblog/pkg/third/qiniu"
	"github.com/eiblog/eiblog/tools"

	"github.com/sirupsen/logrus"
)

var (
	XMLTemplate  *template.Template // template/xml模板
	HTMLTemplate *template.Template // website/html模板

	Store           store.Store // 数据库存储
	Ei              *Cache      // 博客数据缓存
	TwoFactorSecret string      // 缓存两步验证密钥

	ESClient     *es.ESClient         // es 客户端
	DisqusClient *disqus.DisqusClient // disqus 客户端
	QiniuClient  *qiniu.QiniuClient   // qiniu客户端
	Pinger       *pinger.Pinger       // pinger 客户端
)

func init() {
	var err error
	tools.TimeLocation, err = time.LoadLocation(config.Conf.General.Timezone)
	if err != nil {
		logrus.Fatal("init timezone: ", err)
	}
	// 模板解析初始化
	root := filepath.Join(config.EtcDir, "xml", "*.xml")
	XMLTemplate, err = template.New("eiblog").Funcs(tools.TplFuncMap).ParseGlob(root)
	if err != nil {
		logrus.Fatal("init xml template: ", err)
	}
	root = filepath.Join(config.EtcDir, "template")
	files := tools.ReadDirFiles(root, func(fi fs.DirEntry) bool {
		// should not read dir & .DS_Store
		return strings.HasPrefix(fi.Name(), ".")
	})
	HTMLTemplate, err = template.New("eiblog").Funcs(tools.TplFuncMap).ParseFiles(files...)
	if err != nil {
		logrus.Fatal("init html template: ", err)
	}

	// 数据库初始化
	logrus.Info("store drivers: ", store.Drivers())
	Store, err = store.NewStore(config.Conf.Database)
	if err != nil {
		logrus.Fatal("init store: ", err)
	}

	Ei, err = NewCache()
	if err != nil {
		logrus.Fatal("init blog cache: ", err)
	}

	if config.Conf.ESHost != "" {
		ESClient, err = es.NewESClient(config.Conf.ESHost)
		if err != nil {
			logrus.Fatal("init es client: ", err)
		}
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

	// 启动定时器
	go startTimer()
}
