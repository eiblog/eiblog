package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/eiblog/eiblog/pkg/config"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var (
	// Conf 配置
	Conf Config
	// EtcDir 工作目录
	EtcDir string
)

// Config config
type Config struct {
	config.APIMode
	// 静态资源版本, 每次更改了 js/css 都需要提高该值
	StaticVersion int
	// 数据库配置
	Database config.Database
	// 热词, 手动指定, 用于搜索
	HotWords []string
	// Elasticsearch 配置
	ESHost string

	General config.General
	Disqus  config.Disqus
	Google  config.Google
	Qiniu   config.Qiniu
	Twitter config.Twitter
	FeedRPC config.FeedRPC
	Account config.Account
	Pages   []config.CustomPage
}

// init 初始化配置
func init() {
	// run mode
	mode := config.RunModeProd
	if rm := os.Getenv("RUN_MODE"); rm != "" {
		mode = config.RunMode(rm)
		if !mode.IsRunMode() {
			panic("config: unsupported env RUN_MODE: " + mode)
		}
	}
	logrus.Infof("Run mode:%s", mode)

	// 加载配置文件
	var err error
	EtcDir, err = config.WorkEtcPath()
	if err != nil {
		panic(err)
	}
	path := filepath.Join(EtcDir, "app.yml")

	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(data, &Conf)
	if err != nil {
		panic(err)
	}
	Conf.RunMode = mode

	// 读取环境变量配置
	readDatabaseEnv()
}

func readDatabaseEnv() {
	key := strings.ToUpper(Conf.Name) + "_DB_DRIVER"
	if d := os.Getenv(key); d != "" {
		Conf.Database.Driver = d
	}
	key = strings.ToUpper(Conf.Name) + "_DB_SOURCE"
	if s := os.Getenv(key); s != "" {
		Conf.Database.Source = s
	}
}
