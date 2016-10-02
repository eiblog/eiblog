// Package setting provides ...
package setting

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/eiblog/utils/logd"
	"gopkg.in/yaml.v2"
)

const (
	DEV  = "dev"
	PROD = "prod"
)

var (
	wd, _   = os.Getwd()
	Conf    = new(Config)
	BlackIP = make(map[string]bool)
)

type Config struct {
	StaticVersion int      // 当前静态文件版本
	RunMode       string   // 运行模式
	Trash         int      // 回收箱文章保留时间
	Clean         int      // 清理回收箱频率
	PageNum       int      // 前端每页文章数量
	PageSize      int      // 后台每页文章数量
	Length        int      // 自动截取预览长度
	Identifier    string   // 截取标示
	Favicon       string   // icon地址
	StartID       int32    // 文章起始id
	Static        string   // cdn地址
	SearchURL     string   // elasticsearch 地址
	Disqus        struct { // 获取文章数量相关
		ShortName string
		PublicKey string
		URL       string
		Interval  int
	}
	Mode    RunMode  // 运行模式
	Twitter string   // twitter地址
	Blogger struct { // 初始化数据
		BlogName  string
		SubTitle  string
		BeiAn     string
		BTitle    string
		Copyright string
	}
	Account struct {
		Username    string // *
		Password    string // *
		Email       string
		PhoneNumber string
		Address     string
	}
}

type RunMode struct {
	EnableHttp  bool
	HttpPort    int
	EnableHttps bool
	HttpsPort   int
	CertFile    string
	KeyFile     string
	Domain      string
}

func init() {
	// 初始化配置
	dir := wd + "/conf"
	data, err := ioutil.ReadFile(path.Join(dir, "app.yml"))
	checkError(err)
	err = yaml.Unmarshal(data, Conf)
	checkError(err)

	data, err = ioutil.ReadFile(path.Join(dir, "blackip.yml"))
	checkError(err)
	err = yaml.Unmarshal(data, BlackIP)
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		logd.Fatal(err)
	}
}
